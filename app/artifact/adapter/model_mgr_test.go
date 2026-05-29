package adapter

import (
	"testing"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
)

func TestSplitNameExt(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		wantStem  string
		wantExt   string
	}{
		// 单扩展（first-dot / last-dot 等价）
		{"single ext", "foo.txt", "foo", ".txt"},
		{"no ext", "foo", "foo", ""},
		{"trailing dot keeps empty ext", "foo.", "foo", "."},

		// 复合扩展白名单 → first-dot
		{"tar.gz hits whitelist", "foo.tar.gz", "foo", ".tar.gz"},
		{"tar.bz2 hits whitelist", "foo.tar.bz2", "foo", ".tar.bz2"},
		{"tar.xz hits whitelist", "foo.tar.xz", "foo", ".tar.xz"},
		{"tar.zst hits whitelist", "foo.tar.zst", "foo", ".tar.zst"},
		{"min.js hits whitelist", "bundle.min.js", "bundle", ".min.js"},
		{"min.css hits whitelist", "site.min.css", "site", ".min.css"},
		{"d.ts hits whitelist", "types.d.ts", "types", ".d.ts"},
		{"spec.ts hits whitelist", "user.spec.ts", "user", ".spec.ts"},
		{"test.ts hits whitelist", "user.test.ts", "user", ".test.ts"},
		{"min.js precedes .js", "a.b.min.js", "a.b", ".min.js"},
		{"case-insensitive whitelist", "Foo.TAR.GZ", "Foo", ".TAR.GZ"},

		// 非白名单复合点 → last-dot 回落（避免误判）
		{"dotted version not whitelisted", "version.1.2.3.txt", "version.1.2.3", ".txt"},

		// 隐藏文件
		{"leading dot is not ext", ".bashrc", ".bashrc", ""},
		{"hidden file with whitelist suffix still resolves", ".hidden.tar.gz", ".hidden", ".tar.gz"},

		// unicode
		{"unicode name with whitelist", "中文.tar.gz", "中文", ".tar.gz"},
		{"unicode name without whitelist", "中文.zip", "中文", ".zip"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stem, ext := splitNameExt(tc.input)
			if stem != tc.wantStem || ext != tc.wantExt {
				t.Fatalf("splitNameExt(%q) = (%q, %q), want (%q, %q)",
					tc.input, stem, ext, tc.wantStem, tc.wantExt)
			}
		})
	}
}

func TestIsSameContent(t *testing.T) {
	checksum := (&Hash{
		Md5:    "m",
		Sha1:   "s1",
		Sha256: "s256",
		Sha512: "s512",
	}).String()

	t.Run("all three hashes equal -> duplicate", func(t *testing.T) {
		asset := &model.ArtifactAsset{CheckSum: checksum}
		desc := &AssetDescriptor{Hash: &Hash{Md5: "m", Sha1: "s1", Sha256: "s256"}}
		if !isSameContent(asset, desc) {
			t.Fatalf("expected duplicate to be detected when md5/sha1/sha256 all match")
		}
	})

	t.Run("any hash differs -> not duplicate", func(t *testing.T) {
		asset := &model.ArtifactAsset{CheckSum: checksum}
		desc := &AssetDescriptor{Hash: &Hash{Md5: "m", Sha1: "s1", Sha256: "DIFFERENT"}}
		if isSameContent(asset, desc) {
			t.Fatalf("expected sha256 mismatch to defeat duplicate detection")
		}
	})

	t.Run("missing component in stored checksum -> not duplicate", func(t *testing.T) {
		partial := (&Hash{Md5: "m", Sha1: "s1"}).String() // sha256 empty
		asset := &model.ArtifactAsset{CheckSum: partial}
		desc := &AssetDescriptor{Hash: &Hash{Md5: "m", Sha1: "s1", Sha256: "s256"}}
		if isSameContent(asset, desc) {
			t.Fatalf("expected partial stored checksum to be treated as non-duplicate")
		}
	})

	t.Run("nil descriptor hash -> not duplicate", func(t *testing.T) {
		asset := &model.ArtifactAsset{CheckSum: checksum}
		desc := &AssetDescriptor{}
		if isSameContent(asset, desc) {
			t.Fatalf("nil descriptor hash must never be treated as duplicate")
		}
	})

	t.Run("garbage checksum -> not duplicate", func(t *testing.T) {
		asset := &model.ArtifactAsset{CheckSum: "not-json"}
		desc := &AssetDescriptor{Hash: &Hash{Md5: "m", Sha1: "s1", Sha256: "s256"}}
		if isSameContent(asset, desc) {
			t.Fatalf("invalid checksum payload must fall through to non-duplicate")
		}
	})
}

func TestResolveAssetPath(t *testing.T) {
	mm := &ModelManager{}

	t.Run("first slot wins when nothing taken", func(t *testing.T) {
		got, err := mm.resolveAssetPath("foo.tar.gz", func(string) (bool, error) { return false, nil })
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if got != "foo(1).tar.gz" {
			t.Fatalf("expected foo(1).tar.gz, got %q", got)
		}
	})

	t.Run("skip taken slots and use first free", func(t *testing.T) {
		taken := map[string]bool{
			"foo(1).txt": true,
			"foo(2).txt": true,
		}
		got, err := mm.resolveAssetPath("foo.txt", func(p string) (bool, error) {
			return taken[p], nil
		})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if got != "foo(3).txt" {
			t.Fatalf("expected foo(3).txt, got %q", got)
		}
	})

	t.Run("no extension still gets suffix", func(t *testing.T) {
		got, err := mm.resolveAssetPath("README", func(string) (bool, error) { return false, nil })
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if got != "README(1)" {
			t.Fatalf("expected README(1), got %q", got)
		}
	})
}
