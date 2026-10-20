import SwiftUI

@main
struct RummyAssistantApp: App {
    var body: some Scene {
        WindowGroup { OfflineAssistantView().ignoresSafeArea(.container, edges: .bottom) }
    }
}
