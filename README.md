# Fruit Fury — Ebitengine Android + Desktop

An original fruit-slicing arcade game written in **Go with Ebitengine 2.10.5**. This source tree is adapted for **Android touchscreen play**, while keeping the original desktop game working.

**Important:** This ZIP is the buildable *source project*, **not an APK**. The provided GitHub Actions workflow or local build script produces an installable, debug-signed APK. A real Android APK build was **not possible in the environment used to prepare this project** (Go 1.23, no Android SDK/NDK, no dependency network access). Consequently, Android compilation and device compatibility have not been verified here. The core game logic and input-selection tests pass against local API stubs; actual Ebitengine/Android integration requires a full build.

## Android features

- Swipe one finger across fruit to slice, with stable finger tracking across frames; multiple fingers won't create an accidental long slash.
- Tap **START SLICING**, **RESUME GAME**, **PLAY AGAIN**, **BACK TO MENU**, or the onscreen **pause** control.
- Tap the **SOUND ON/OFF** button at the bottom-left during gameplay to mute/unmute.
- Landscape full-screen display, with 1280×720 logical game coordinates scaled by Ebitengine.
- Pauses rendering when sent to the background, resumes on return.
- Persists high scores in the app's private files directory (no storage permission).
- Original gameplay preserved: 60-second rounds, three lives, fruit/bombs/golden fruit, combos, juice particles, split halves, synthesized sound, best scores.
- Includes custom launcher icon. No permission to access the internet, contacts, or shared storage.

## Easiest APK build: GitHub Actions

This is useful if you don't have Android Studio or Android development tools installed locally.

1. Create a **new GitHub repository** and upload all contents of this project (including the hidden `.github` folder). Don't upload only the ZIP itself.
2. Go to **Actions → Build Fruit Fury Android APK → Run workflow**. If Actions asks you to enable workflows, enable them for your repository.
3. Wait for the `Android APK (ARM64)` job to finish.
4. Open the successful run and download its **Fruit-Fury-Android-Debug-APK** artifact under *Artifacts*.
5. Extract that downloaded artifact ZIP. Its contents are `fruit-fury-android-debug.apk`. Transfer this APK to an ARM64 Android phone and open it to install. Android may ask you to allow installs from that file manager/browser.

Workflow file: [`.github/workflows/android-apk.yml`](.github/workflows/android-apk.yml). The workflow downloads the dependencies using a normal GitHub-hosted Ubuntu runner. It uses Java 17, Go 1.25, Android API 35 and NDK r27c, Gradle 8.10.2, runs the build script, and uploads the debug APK artifact. **Not executed in this environment.**

> The resulting APK is **debug-signed**, suitable for sideloading/testing. It is **not** a Play Store release build. For store distribution, add a secure release signing configuration and follow Play Console requirements.

## Build APK locally

Required development tools:

- **Go 1.25 or newer** (the `go.mod` requires Go 1.25)
- Java **JDK 17** (the Gradle Android plugin uses JDK 17)
- Android SDK with **`platforms;android-35`** and **`build-tools;35.0.0`**
- Android NDK **`27.2.12479018`** (r27c)
- **Gradle 8.10.2** available as `gradle` in PATH
- Internet access on first build to fetch Go modules, Gradle plugins, Android components

If the Android SDK command-line tools (`sdkmanager`) are installed, install the required components:

```sh
sdkmanager --install 'platforms;android-35' 'build-tools;35.0.0' 'ndk;27.2.12479018'
```

Set `ANDROID_HOME` (or `ANDROID_SDK_ROOT`) to your SDK path. Then from the repository's top-level directory:

```sh
./scripts/build-android.sh
```

This script:

1. Downloads/install `ebitenmobile` v2.10.5 if necessary.
2. Resolves the project Go modules (`go mod tidy`).
3. Runs `ebitenmobile bind -target android/arm64 -androidapi 23 -javapkg com.fruitfury.game -o android/app/libs/fruitfury.aar ./mobile`.
4. Runs Gradle `:app:assembleDebug`.
5. Copies the result to **`dist/fruit-fury-android-debug.apk`**.

APK package/application ID: **`com.fruitfury.game`**. Minimum Android version: **6.0 (API 23)**. The bundled native library targets **ARM64** (most modern Android phones). For older 32-bit ARM phones, modify the `-target` flag and rebuild; for emulator use, build for the emulator's architecture.

## Desktop build still works

```sh
go mod tidy
go run .
```

Desktop controls: hold left mouse and drag to slice; `P`/`Esc` pause; `M` mute; `R` restart after game over. The same touch input works on Android; no keyboard is required there.

## Project layout

| Directory/file | Role |
|---|---|
| `internal/game/` | Shared Go game, rendering, sound, scoring, touch+mouse input, tests |
| `main.go` | Original desktop launcher |
| `mobile/mobile.go` | Ebitengine mobile binding; exports `SetDataDirectory` for private Android score saves |
| `android/app/src/main/` | Native Java full-screen Activity, Android manifest, app icon |
| `android/app/build.gradle` | Gradle Android app definition consuming the generated `.aar` |
| `scripts/build-android.sh` | Build APK with the real Go/Android toolchain |
| `.github/workflows/android-apk.yml` | Automatically builds an APK and uploads it as a GitHub Actions artifact |
| `art-preview.png` | Illustrative scene, not a runtime screenshot or game dependency |

## Gameplay details

- A round lasts **60 seconds**; missing fruit or slicing a bomb costs one of three lives.
- Sliced fruit awards points. Fast successive slices generate a combo multiplier.
- Golden fruit awards a larger bonus and restores three seconds (up to 75 seconds).
- Procedural art and synthesized sounds require no asset downloads at runtime.

## Verification notes

- `gofmt` completed.
- Go unit tests and `go vet` **passed using local Ebitengine API stubs** (because external modules could not be downloaded here). This checks internal gameplay code; it is not a full-platform compiler check.
- Java Activity syntax checked using locally defined Android API stubs. Manifest XML and GitHub Actions YAML validated.
- **No real APK or Android device test was performed.** This is a ready-to-build Android port, not a verified APK binary.

Framework: [Ebitengine](https://ebitengine.org/en/documents/mobile.html), licensed Apache 2.0. Game design, synthesized sound, and graphics are original to this project. Included fonts come from Go's `golang.org/x/image` module.
