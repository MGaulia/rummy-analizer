import SwiftUI
import WebKit
import PhotosUI
import UniformTypeIdentifiers

// Bundled file URLs and embedded Wasm bytes need no localhost server. No app
// network permission, remote URL, or photo upload path is part of this bridge.
struct OfflineAssistantView: UIViewRepresentable {
    func makeCoordinator() -> Coordinator { Coordinator() }
    func makeUIView(context: Context) -> WKWebView {
        let configuration = WKWebViewConfiguration()
        configuration.userContentController.add(context.coordinator, name: "photo")
        let webView = WKWebView(frame: .zero, configuration: configuration)
        context.coordinator.webView = webView
        webView.navigationDelegate = context.coordinator
        webView.isInspectable = false
        if let root = Bundle.main.url(forResource: "web", withExtension: nil),
           FileManager.default.fileExists(atPath: root.appendingPathComponent("engine/wasm_data.js").path) {
            webView.loadFileURL(root.appendingPathComponent("index.html"), allowingReadAccessTo: root)
        } else {
            webView.loadHTMLString("<p>Offline engine assets are missing. Run scripts/build-phone.sh before building the app.</p>", baseURL: nil)
        }
        return webView
    }
    func updateUIView(_ uiView: WKWebView, context: Context) {}
    static func dismantleUIView(_ uiView: WKWebView, coordinator: Coordinator) {
        uiView.configuration.userContentController.removeScriptMessageHandler(forName: "photo")
        uiView.stopLoading()
    }

    final class Coordinator: NSObject, WKScriptMessageHandler, WKNavigationDelegate,
                             PHPickerViewControllerDelegate, UIImagePickerControllerDelegate,
                             UINavigationControllerDelegate {
        weak var webView: WKWebView?
        private var zone = "table"
        private var pending = false

        func userContentController(_ userContentController: WKUserContentController, didReceive message: WKScriptMessage) {
            guard !pending, message.frameInfo.isMainFrame,
                  let body = message.body as? [String: Any],
                  let requestedZone = body["zone"] as? String,
                  ["table", "rack"].contains(requestedZone),
                  let presenter = presenter() else { return }
            zone = requestedZone
            if body["camera"] as? Bool == true {
                guard UIImagePickerController.isSourceTypeAvailable(.camera) else {
                    send(["error": "No camera is available. Choose an existing photo."]); return
                }
                let picker = UIImagePickerController()
                picker.sourceType = .camera; picker.delegate = self
                pending = true; presenter.present(picker, animated: true)
            } else {
                var config = PHPickerConfiguration(photoLibrary: .shared())
                config.filter = .images; config.selectionLimit = 1
                let picker = PHPickerViewController(configuration: config)
                picker.delegate = self
                pending = true; presenter.present(picker, animated: true)
            }
        }

        private func presenter() -> UIViewController? {
            guard let scene = webView?.window?.windowScene,
                  var controller = scene.windows.first(where: { $0.isKeyWindow })?.rootViewController else { return nil }
            while let presented = controller.presentedViewController { controller = presented }
            return controller
        }
        func picker(_ picker: PHPickerViewController, didFinishPicking results: [PHPickerResult]) {
            picker.dismiss(animated: true)
            guard let result = results.first else { pending = false; return }
            guard result.itemProvider.canLoadObject(ofClass: UIImage.self) else {
                pending = false; send(["error": "Choose a readable image."]); return
            }
            let selectedZone = zone
            result.itemProvider.loadObject(ofClass: UIImage.self) { [weak self] object, error in
                guard let image = object as? UIImage else {
                    DispatchQueue.main.async { self?.pending = false; self?.send(["error": error?.localizedDescription ?? "Could not read this photo."]) }; return
                }
                self?.recognize(image, zone: selectedZone)
            }
        }
        func imagePickerController(_ picker: UIImagePickerController, didFinishPickingMediaWithInfo info: [UIImagePickerController.InfoKey: Any]) {
            picker.dismiss(animated: true)
            guard let image = info[.originalImage] as? UIImage else { pending = false; return }
            recognize(image, zone: zone)
        }
        func imagePickerControllerDidCancel(_ picker: UIImagePickerController) { pending = false; picker.dismiss(animated: true) }

        private func recognize(_ image: UIImage, zone: String) {
            DispatchQueue.global(qos: .userInitiated).async { [weak self] in
                do {
                    let payload = try TileRecognition.observe(image: image, zone: zone)
                    DispatchQueue.main.async { self?.pending = false; self?.send(payload) }
                } catch {
                    DispatchQueue.main.async { self?.pending = false; self?.send(["error": error.localizedDescription]) }
                }
            }
        }
        private func send(_ payload: [String: Any]) {
            do {
                let data = try JSONSerialization.data(withJSONObject: payload)
                guard var json = String(data: data, encoding: .utf8) else { throw TileRecognition.Failure.image }
                // Serialize data, never interpolate unescaped OCR into code.
                json = json.replacingOccurrences(of: "\u{2028}", with: "\\u2028")
                           .replacingOccurrences(of: "\u{2029}", with: "\\u2029")
                webView?.evaluateJavaScript("void window.acceptNativePhoto(\(json));") { [weak self] _, error in
                    if let error = error { self?.showTransferError(error.localizedDescription) }
                }
            } catch { showTransferError(error.localizedDescription) }
        }
        private func showTransferError(_ message: String) {
            let alert = UIAlertController(title: "Could not read photo", message: message, preferredStyle: .alert)
            alert.addAction(UIAlertAction(title: "OK", style: .default))
            presenter()?.present(alert, animated: true)
        }
        func webView(_ webView: WKWebView, decidePolicyFor navigationAction: WKNavigationAction, decisionHandler: @escaping (WKNavigationActionPolicy) -> Void) {
            guard let url = navigationAction.request.url, url.isFileURL,
                  let root = Bundle.main.url(forResource: "web", withExtension: nil),
                  url.standardizedFileURL.path.hasPrefix(root.standardizedFileURL.path + "/") else {
                decisionHandler(.cancel); return
            }
            decisionHandler(.allow)
        }
    }
}
