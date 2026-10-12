#!/usr/bin/env bash
# Build a sideloadable, debug-signed APK using Ebitengine and Gradle.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

export ANDROID_HOME="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}"
if [[ -z "$ANDROID_HOME" || ! -d "$ANDROID_HOME" ]]; then
    echo 'ERROR: ANDROID_HOME must point to the Android SDK.' >&2
    exit 1
fi

# Android NDK 27 is recommended here. Avoid NDK 30 for older Oboe versions.
if [[ -z "${ANDROID_NDK_HOME:-}" ]]; then
    export ANDROID_NDK_HOME="$ANDROID_HOME/ndk/27.2.12479018"
fi
if [[ ! -d "$ANDROID_NDK_HOME" ]]; then
    echo "ERROR: Android NDK not found: $ANDROID_NDK_HOME" >&2
    exit 1
fi
if ! command -v gradle >/dev/null; then
    echo 'ERROR: Gradle 8.9+ must be installed and available on PATH.' >&2
    exit 1
fi

export PATH="$(go env GOPATH)/bin:$PATH"
if ! command -v ebitenmobile >/dev/null; then
    echo 'Installing the Ebitengine Android binder...'
    go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@v2.10.5
fi

echo 'Resolving Go dependencies...'
go mod tidy
mkdir -p android/app/libs dist

echo 'Binding the Go game for Android ARM64...'
ebitenmobile bind     -target android/arm64     -androidapi 23     -javapkg com.fruitfury.game     -o android/app/libs/fruitfury.aar     ./mobile

echo 'Assembling the debug-signed Android APK...'
(cd android && gradle --no-daemon :app:assembleDebug)

out=dist/fruit-fury-android-debug.apk
cp android/app/build/outputs/apk/debug/app-debug.apk "$out"
echo "APK created: $root/$out"
