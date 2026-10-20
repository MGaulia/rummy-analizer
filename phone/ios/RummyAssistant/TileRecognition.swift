import UIKit
import Vision
import CryptoKit

// Prototype local detector/OCR pipeline: Apple's installed Vision framework,
// rectangle proposals, and glyph-color suggestions. No model download or API.
// All proposals require explicit correction/confirmation; detection confidence
// is not an accuracy measurement. Real-photo/device evaluation is outstanding.
enum TileRecognition {
    enum Failure: LocalizedError {
        case image
        var errorDescription: String? { "The photo could not be decoded. Try a clearer image." }
    }

    static func observe(image: UIImage, zone: String) throws -> [String: Any] {
        // Normalize UIImage orientation and bound memory/time before Vision.
        let scale = min(1, 2000 / max(image.size.width, image.size.height))
        let size = CGSize(width: max(1, image.size.width * scale), height: max(1, image.size.height * scale))
        let format = UIGraphicsImageRendererFormat(); format.scale = 1
        let normalized = UIGraphicsImageRenderer(size: size, format: format).image { _ in image.draw(in: CGRect(origin: .zero, size: size)) }
        guard let cg = normalized.cgImage, let jpeg = normalized.jpegData(compressionQuality: 0.9) else { throw Failure.image }
        let source = SHA256.hash(data: jpeg).map { String(format: "%02x", $0) }.joined()
        let rectangles = VNDetectRectanglesRequest()
        rectangles.revision = VNDetectRectanglesRequestRevision1
        rectangles.minimumAspectRatio = 0.3; rectangles.maximumAspectRatio = 1.0
        rectangles.minimumConfidence = 0.35; rectangles.minimumSize = 0.018
        rectangles.quadratureTolerance = 30; rectangles.maximumObservations = 180
        try VNImageRequestHandler(cgImage: cg, options: [:]).perform([rectangles])
        // Convert to top-left normalized coordinates. Row grouping is only a
        // suggestion; users confirm actual group boundaries and order.
        var boxes = (rectangles.results ?? []).map { r -> CGRect in
            CGRect(x: r.boundingBox.minX, y: 1 - r.boundingBox.maxY,
                   width: r.boundingBox.width, height: r.boundingBox.height)
        }
        boxes.sort { $0.minY < $1.minY }
        var rows: [[CGRect]] = []
        for box in boxes {
            if let index = rows.firstIndex(where: { row in
                guard let first = row.first else { return false }
                return abs(box.midY - first.midY) < min(box.height, first.height) * 0.45
            }) { rows[index].append(box) } else { rows.append([box]) }
        }
        var observations: [[String: Any]] = []
        var meld = 0
        for row in rows {
            var previous: CGRect?
            var order = 0
            for box in row.sorted(by: { $0.minX < $1.minX }) {
                if previous == nil || box.minX - previous!.maxX > min(box.width, previous!.width) * 0.5 { meld += 1; order = 0 }
                order += 1; previous = box
                let pixels = CGRect(x: box.minX * Double(cg.width), y: box.minY * Double(cg.height),
                                    width: box.width * Double(cg.width), height: box.height * Double(cg.height)).integral
                guard let crop = cg.cropping(to: pixels.intersection(CGRect(x: 0, y: 0, width: cg.width, height: cg.height))) else { continue }
                let textRequest = VNRecognizeTextRequest()
                textRequest.revision = VNRecognizeTextRequestRevision3
                textRequest.recognitionLevel = .accurate; textRequest.usesLanguageCorrection = false
                textRequest.recognitionLanguages = ["en-US"]
                try VNImageRequestHandler(cgImage: crop, options: [:]).perform([textRequest])
                let textResults = textRequest.results ?? []
                var candidates: [[String: Any]] = []
                var texts: [String] = []
                for observation in textResults {
                    for text in observation.topCandidates(3) {
                        let value = text.string.trimmingCharacters(in: .whitespacesAndNewlines)
                        texts.append(value)
                        if let n = Int(value), (1...13).contains(n), value == String(n) {
                            let glyph = observation.boundingBox
                            let glyphRect = CGRect(x: glyph.minX * Double(crop.width), y: (1 - glyph.maxY) * Double(crop.height), width: glyph.width * Double(crop.width), height: glyph.height * Double(crop.height))
                            if let glyphImage = crop.cropping(to: glyphRect.integral), let color = suggestColor(glyphImage) {
                                candidates.append(["tile": ["color": color, "number": n], "confidence": Double(text.confidence)])
                            }
                        } else if value.lowercased().contains("joker") {
                            candidates.append(["tile": ["joker": true], "confidence": Double(text.confidence)])
                        }
                    }
                }
                observations.append([
                    "rect": ["x": box.minX, "y": box.minY, "width": box.width, "height": box.height],
                    "text": texts.joined(separator: " / "), "candidates": candidates,
                    "suggestedMeld": meld, "suggestedOrder": order
                ])
            }
        }
        return ["zone": zone, "source": source, "image": "data:image/jpeg;base64," + jpeg.base64EncodedString(),
                "observations": observations,
                "recognizer": ["name": "apple-vision-rectangle-text-color", "pipelineRevision": 1,
                               "osVersion": UIDevice.current.systemVersion, "maxDimension": 2000,
                               "rectangleRevision": 1, "textRevision": 3],
                "notice": "Local Vision suggestions are uncalibrated. Check every color, number, joker, missed tile, meld boundary, and tile order before confirming."]
    }

    // Suggested physical colors only, never confirmed game values. Discard
    // pale tile background pixels; classify colored/dark glyph pixels.
    private static func suggestColor(_ image: CGImage) -> String? {
        let width = 32, height = 32
        var rgba = [UInt8](repeating: 0, count: width * height * 4)
        let space = CGColorSpaceCreateDeviceRGB()
        let counts: [String: Int]? = rgba.withUnsafeMutableBytes { buffer in
            guard let context = CGContext(data: buffer.baseAddress, width: width, height: height, bitsPerComponent: 8,
                                          bytesPerRow: width * 4, space: space, bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue) else { return nil }
            context.draw(image, in: CGRect(x: 0, y: 0, width: width, height: height))
            var counts = ["black": 0, "blue": 0, "orange": 0, "red": 0]
            let bytes = buffer.bindMemory(to: UInt8.self)
            for i in stride(from: 0, to: bytes.count, by: 4) {
                let r = Double(bytes[i]) / 255, g = Double(bytes[i + 1]) / 255, b = Double(bytes[i + 2]) / 255
                let high = max(r, max(g, b)), low = min(r, min(g, b))
                if high < 0.36 { counts["black", default: 0] += 1; continue }
                if high - low < 0.18 { continue }
                if b > r * 1.15 && b > g * 1.08 { counts["blue", default: 0] += 1 }
                else if r > g * 1.12 && r > b * 1.25 { counts[g > r * 0.38 ? "orange" : "red", default: 0] += 1 }
            }
            return counts
        }
        guard let counts = counts, let best = counts.max(by: { $0.value < $1.value }), best.value > 4 else { return nil }
        return best.key
    }
}
