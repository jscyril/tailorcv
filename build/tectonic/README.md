# Pinned Tectonic runtime

TailorCV packages Tectonic 0.16.9 from the official platform archives and verifies each archive with its release SHA-256 digest. `cmd/packagetectonic` hydrates the exact TeX Live 2022.0r0 files required by the built-in templates, writes them to a compact local ZIP bundle, and verifies both fixtures with Tectonic's `--only-cached` mode and unusable network proxies.

The executable and `tectonic-resources.zip` are placed together in a `bin` directory beside the application executable. The runtime compiler requires the local bundle and always passes `--only-cached`, so ordinary resume compilation cannot fetch missing TeX packages. A custom template that needs a package outside this curated resource set fails with a compiler diagnostic instead of accessing the network. Because adding files changes a macOS application bundle, CI restores and verifies Wails' ad-hoc signature after packaging.

From the repository root, package the current platform into a chosen runtime directory:

```bash
go run ./cmd/packagetectonic -destination /path/to/application/bin
```

The platform builds in `.github/workflows/ci.yml` run this command after the Wails build and then execute the real compiler integration test and disposable packaged application workflow against the executable and resource bundle.
