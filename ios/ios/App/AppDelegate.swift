import UIKit
import Core

@UIApplicationMain
class AppDelegate: UIResponder, UIApplicationDelegate {
    let core = CoreCore("C02CK1XKMD6R.local:9001", username: "Seb")
    var window: UIWindow?

    private func setupWindow() {
        // Create the window (bypass the storyboard)
        window = UIWindow(frame: UIScreen.main.bounds)

        let rootVC = UIViewController()

        // Force dark mode for iOS <= 13.0
        if #available(iOS 13.0, *) {
            window?.overrideUserInterfaceStyle = .dark
        }

        window?.makeKeyAndVisible()
        window?.rootViewController = rootVC

    }

    func applicationDidFinishLaunching(_ application: UIApplication) {
        setupWindow()

        do {
            try core?.start()
        } catch let error {
            print(error)
        }

        Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { [weak self] timer in
            guard let self = self else { return }
            self.core?.sendMessage("Hello! \(timer.timeInterval)")
        }
    }

    func applicationWillTerminate(_ application: UIApplication) {
        core?.stop()
    }
}
