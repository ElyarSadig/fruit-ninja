#!/usr/bin/env bash
# Build a debug-signed Android APK using Ebitengine's Android AAR binding.
# When run in GitHub Actions, the calling workflow saves this output to build.log.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

export ANDROID_HOME="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}"
if [[ -z "$ANDROID_HOME" || ! -d "$ANDROID_HOME" ]]; then
    echo 'ERROR: ANDROID_HOME or ANDROID_SDK_ROOT must point to the Android SDK.' >&2
    exit 1
fi

# Use NDK r27c instead of r30, which has an incompatibility with older Oboe headers.
export ANDROID_NDK_HOME="${ANDROID_NDK_HOME:-$ANDROID_HOME/ndk/27.2.12479018}"
if [[ ! -d "$ANDROID_NDK_HOME" ]]; then
    echo "ERROR: Android NDK not found: $ANDROID_NDK_HOME" >&2
    exit 1
fi
if ! command -v gradle >/dev/null 2>&1; then
    echo 'ERROR: Gradle must be installed and on PATH.' >&2
    exit 1
fi

export PATH="$(go env GOPATH)/bin:$PATH"
if ! command -v ebitenmobile >/dev/null 2>&1; then
    echo '== Installing ebitenmobile v2.10.5 =='
    go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@v2.10.5
fi

echo '== Resolving Go dependencies =='
go mod tidy
mkdir -p android/app/libs dist

echo '== Binding Go game for Android ARM64 =='
ebitenmobile bind \
    -target android/arm64 \
    -androidapi 23 \
    -javapkg com.fruitfury.game \
    -v \
    -o android/app/libs/fruitfury.aar \
    ./mobile

echo '== Assembling Android debug APK =='
(cd android && gradle --no-daemon :app:assembleDebug --stacktrace)

out=dist/fruit-fury-android-debug.apk
cp android/app/build/outputs/apk/debug/app-debug.apk "$out"
test -s "$out"
echo "APK created: $root/$out"
