package storage

import (
	"regexp"
	"strings"
	"testing"

	"github.com/MTudorrrr/distribution"
	"github.com/MTudorrrr/distribution/internal/dcontext"
	"github.com/MTudorrrr/distribution/manifest"
	"github.com/MTudorrrr/distribution/manifest/schema2"
	"github.com/MTudorrrr/distribution/registry/storage/driver/inmemory"
	"github.com/opencontainers/go-digest"
)

func TestVerifyManifestForeignLayer(t *testing.T) {
	ctx := dcontext.Background()
	inmemoryDriver := inmemory.New()
	registry := createRegistry(t, inmemoryDriver,
		ManifestURLsAllowRegexp(regexp.MustCompile("^https?://foo")),
		ManifestURLsDenyRegexp(regexp.MustCompile("^https?://foo/nope")))
	repo := makeRepository(t, registry, "test")
	manifestService := makeManifestService(t, repo)

	config, err := repo.Blobs(ctx).Put(ctx, schema2.MediaTypeImageConfig, nil)
	if err != nil {
		t.Fatal(err)
	}

	layer, err := repo.Blobs(ctx).Put(ctx, schema2.MediaTypeLayer, nil)
	if err != nil {
		t.Fatal(err)
	}

	foreignLayer := distribution.Descriptor{
		Digest:    "sha256:463435349086340864309863409683460843608348608934092322395278926a",
		Size:      6323,
		MediaType: schema2.MediaTypeForeignLayer,
	}

	emptyLayer := distribution.Descriptor{
		Digest: "",
	}

	template := schema2.Manifest{
		Versioned: manifest.Versioned{
			SchemaVersion: 2,
			MediaType:     schema2.MediaTypeManifest,
		},
		Config: config,
	}

	type testcase struct {
		BaseLayer distribution.Descriptor
		URLs      []string
		Err       error
	}

	cases := []testcase{
		{
			foreignLayer,
			nil,
			errMissingURL,
		},
		{
			// regular layers may have foreign urls
			layer,
			[]string{"http://foo/bar"},
			nil,
		},
		{
			foreignLayer,
			[]string{"file:///local/file"},
			errInvalidURL,
		},
		{
			foreignLayer,
			[]string{"http://foo/bar#baz"},
			errInvalidURL,
		},
		{
			foreignLayer,
			[]string{""},
			errInvalidURL,
		},
		{
			foreignLayer,
			[]string{"https://foo/bar", ""},
			errInvalidURL,
		},
		{
			foreignLayer,
			[]string{"", "https://foo/bar"},
			errInvalidURL,
		},
		{
			foreignLayer,
			[]string{"http://nope/bar"},
			errInvalidURL,
		},
		{
			foreignLayer,
			[]string{"http://foo/nope"},
			errInvalidURL,
		},
		{
			foreignLayer,
			[]string{"http://foo/bar"},
			nil,
		},
		{
			foreignLayer,
			[]string{"https://foo/bar"},
			nil,
		},
		{
			emptyLayer,
			[]string{"https://foo/empty"},
			digest.ErrDigestInvalidFormat,
		},
		{
			emptyLayer,
			[]string{},
			digest.ErrDigestInvalidFormat,
		},
	}

	for _, c := range cases {
		m := template
		l := c.BaseLayer
		l.URLs = c.URLs
		m.Layers = []distribution.Descriptor{l}
		dm, err := schema2.FromStruct(m)
		if err != nil {
			t.Error(err)
			continue
		}

		_, err = manifestService.Put(ctx, dm)
		if verr, ok := err.(distribution.ErrManifestVerification); ok {
			// Extract the first error
			if len(verr) == 2 {
				if _, ok = verr[1].(distribution.ErrManifestBlobUnknown); ok {
					err = verr[0]
				}
			}
		}
		if err != c.Err {
			t.Errorf("%#v: expected %v, got %v", l, c.Err, err)
		}
	}
}

func TestVerifyManifestBlobLayerAndConfig(t *testing.T) {
	ctx := dcontext.Background()
	inmemoryDriver := inmemory.New()
	registry := createRegistry(t, inmemoryDriver,
		ManifestURLsAllowRegexp(regexp.MustCompile("^https?://foo")),
		ManifestURLsDenyRegexp(regexp.MustCompile("^https?://foo/nope")))

	repo := makeRepository(t, registry, strings.ToLower(t.Name()))
	manifestService := makeManifestService(t, repo)

	config, err := repo.Blobs(ctx).Put(ctx, schema2.MediaTypeImageConfig, nil)
	if err != nil {
		t.Fatal(err)
	}

	layer, err := repo.Blobs(ctx).Put(ctx, schema2.MediaTypeLayer, nil)
	if err != nil {
		t.Fatal(err)
	}

	template := schema2.Manifest{
		Versioned: manifest.Versioned{
			SchemaVersion: 2,
			MediaType:     schema2.MediaTypeManifest,
		},
	}

	checkFn := func(m schema2.Manifest, rerr error) {
		dm, err := schema2.FromStruct(m)
		if err != nil {
			t.Error(err)
			return
		}

		_, err = manifestService.Put(ctx, dm)
		if verr, ok := err.(distribution.ErrManifestVerification); ok {
			// Extract the first error
			if len(verr) == 2 {
				if _, ok = verr[1].(distribution.ErrManifestBlobUnknown); ok {
					err = verr[0]
				}
			} else if len(verr) == 1 {
				err = verr[0]
			}
		}
		if err != rerr {
			t.Errorf("%#v: expected %v, got %v", m, rerr, err)
		}
	}

	type testcase struct {
		Desc distribution.Descriptor
		URLs []string
		Err  error
	}

	layercases := []testcase{
		// empty media type
		{
			distribution.Descriptor{},
			[]string{"http://foo/bar"},
			digest.ErrDigestInvalidFormat,
		},
		{
			distribution.Descriptor{},
			nil,
			digest.ErrDigestInvalidFormat,
		},
		// unknown media type, but blob is present
		{
			distribution.Descriptor{
				Digest: layer.Digest,
			},
			nil,
			nil,
		},
		{
			distribution.Descriptor{
				Digest: layer.Digest,
			},
			[]string{"http://foo/bar"},
			nil,
		},
		// gzip layer, but invalid digest
		{
			distribution.Descriptor{
				MediaType: schema2.MediaTypeLayer,
			},
			nil,
			digest.ErrDigestInvalidFormat,
		},
		{
			distribution.Descriptor{
				MediaType: schema2.MediaTypeLayer,
			},
			[]string{"https://foo/bar"},
			digest.ErrDigestInvalidFormat,
		},
		{
			distribution.Descriptor{
				MediaType: schema2.MediaTypeLayer,
				Digest:    digest.Digest("invalid"),
			},
			nil,
			digest.ErrDigestInvalidFormat,
		},
		// normal uploaded gzip layer
		{
			layer,
			nil,
			nil,
		},
		{
			layer,
			[]string{"https://foo/bar"},
			nil,
		},
	}

	for _, c := range layercases {
		m := template
		m.Config = config

		l := c.Desc
		l.URLs = c.URLs

		m.Layers = []distribution.Descriptor{l}

		checkFn(m, c.Err)
	}

	configcases := []testcase{
		// valid config
		{
			config,
			nil,
			nil,
		},
		// invalid digest
		{
			distribution.Descriptor{
				MediaType: schema2.MediaTypeImageConfig,
			},
			[]string{"https://foo/bar"},
			digest.ErrDigestInvalidFormat,
		},
		{
			distribution.Descriptor{
				MediaType: schema2.MediaTypeImageConfig,
				Digest:    digest.Digest("invalid"),
			},
			nil,
			digest.ErrDigestInvalidFormat,
		},
	}

	for _, c := range configcases {
		m := template
		m.Config = c.Desc
		m.Config.URLs = c.URLs

		checkFn(m, c.Err)
	}
}
