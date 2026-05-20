# Minimal backend build

This repository carries an opt-in `minimal` build tag for producing a smaller
rclone binary. It is intended for embedded/mobile packaging where only a fixed
set of remotes is required.

## Scope

The `minimal` tag keeps these backends:

- Google Drive (`drive`)
- Microsoft OneDrive (`onedrive`)
- iCloud Drive (`iclouddrive`, service `drive` only)
- Dropbox (`dropbox`)
- WebDAV (`webdav`)
- Local filesystem (`local`)

`local` is kept because rclone uses it for normal local path sources and
destinations. iCloud Photos is deliberately excluded; the iCloud backend only
advertises the `drive` service in minimal builds.

The tag also excludes large optional surfaces that are not needed by the above
set, including the default all-backends registry, plugin loading, RC server
startup, RC flags, self-update, most RC registration code, crypt-specific
`lsjson` handling, and iCloud Photos API code.

## Build a local minimal binary

```sh
GOCACHE=/private/tmp/go-build-cache \
go build \
  -trimpath \
  -tags minimal \
  -ldflags '-s -w -buildid=' \
  -o /private/tmp/rclone-minimal \
  .
```

Check the enabled backends:

```sh
/private/tmp/rclone-minimal help backends
```

Expected backend list:

```text
local
iclouddrive
webdav
dropbox
onedrive
drive
```

Confirm that excluded packages are not linked:

```sh
GOCACHE=/private/tmp/go-build-cache \
go list -deps -tags minimal . |
  rg '^github.com/rclone/rclone/(backend|cmd|fs/rc|vfs|lib/plugin)'
```

The only backend packages should be the five cloud backends above plus `local`
and their direct helper packages. `fs/rc` may still appear, but in a minimal
build it is reduced to `params.go` for shared `rc.Params` helpers; the RC server
and registry are excluded.

## Build for iOS arm64, minimum iOS 15.0

Use the iPhoneOS SDK and set both Go and cgo to target iOS devices:

```sh
IOS_SDK="$(xcrun --sdk iphoneos --show-sdk-path)"
IOS_CLANG="$(xcrun --sdk iphoneos --find clang)"

env \
  GOOS=ios \
  GOARCH=arm64 \
  CGO_ENABLED=1 \
  CC="$IOS_CLANG" \
  CGO_CFLAGS="-isysroot $IOS_SDK -miphoneos-version-min=15.0 -arch arm64" \
  CGO_LDFLAGS="-isysroot $IOS_SDK -miphoneos-version-min=15.0 -arch arm64" \
  GOCACHE=/private/tmp/go-build-cache \
  go build \
    -trimpath \
    -tags minimal \
    -ldflags '-s -w -buildid=' \
    -o /private/tmp/rclone-ios-arm64-minimal \
    .
```

Verify the produced Mach-O:

```sh
file /private/tmp/rclone-ios-arm64-minimal
otool -l /private/tmp/rclone-ios-arm64-minimal |
  rg -A4 'LC_BUILD_VERSION|LC_VERSION_MIN'
```

Expected properties:

- `Mach-O 64-bit executable arm64`
- `platform 2`
- `minos 15.0`

Install into the TrollRecorder app layout:

```sh
install -m 0755 \
  /private/tmp/rclone-ios-arm64-minimal \
  /Users/82flex/Projects/TrollRecorderPrivate/layout/Applications/TRApp.app/rclone
```

The resulting binary is unsigned. Leave signing to the app packaging/signing
pipeline.

