# Homebrew Tap Plan

CEO Harness does not have a published Homebrew tap yet. This is the draft formula shape to use after a real release URL and checksum exist.

`scripts/release-local.sh` also writes a local draft formula to `dist/homebrew/ceo-packet.rb`. That generated file uses the local Darwin archive and checksum so it can be inspected before a tap exists.

```ruby
class CeoHarness < Formula
  desc "Local CEO/subagent coding harness"
  homepage "https://example.invalid/ceo-harness"
  url "https://example.invalid/ceo-packet_0.1.0_darwin_arm64.tar.gz"
  sha256 "<replace-with-release-checksum>"
  version "0.1.0"

  def install
    bin.install "ceo-packet"
  end

  test do
    assert_match "ceo-packet", shell_output("#{bin}/ceo-packet --version")
  end
end
```

Before publishing:

1. Create a public repository and release.
2. Build release archives with `scripts/release-local.sh` or CI.
3. Verify `dist/checksums.txt` from inside `dist/`.
4. Replace the placeholder URL and checksum.
5. Verify the local formula with Homebrew after every placeholder has been replaced.

Do not publish the formula, create a tap, tag, push, or announce remote install support until those placeholders are gone and the remote archive checksum has been rechecked.
