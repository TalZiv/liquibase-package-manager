package main

import (
	"context"
	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/go-version"
	"golang.org/x/oauth2"
	"os"
	"package-manager/internal/app/packages"
	"package-manager/internal/app/utils"
	"sort"
	"strings"
)

type Github struct {}

var client *github.Client
var ctx = context.Background()

func init() {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_PAT")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client = github.NewClient(tc)
}

func (g Github) GetVersions(m Module) []*version.Version {

	rr, _, _ := client.Repositories.ListReleases(context.Background(), m.owner, m.repo, &github.ListOptions{})
	versions := make([]*version.Version, len(rr))
	for i, r := range rr {
		v, _ := version.NewVersion(r.GetTagName())
		versions[i] = v
	}
	sort.Sort(version.Collection(versions))
	return versions
}

func (g Github) GetNewVersions(m Module, p packages.Package) packages.Package {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_PAT")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	for _, v := range m.GetVersions() {
		var ver packages.Version
		ver.Tag = v.Original()
		pv := p.GetVersion(ver.Tag)

		if pv.Tag != "" {
			// if remote version is already in package manifest skip it
			continue
		}

		release, _, _ := client.Repositories.GetReleaseByTag(context.Background(), m.owner, m.repo, ver.Tag)
		for _, a := range release.Assets {
			if strings.HasSuffix(a.GetName(), ".jar") {
				ver.Path = a.GetBrowserDownloadURL()
			}
			if strings.Contains(a.GetName(), "sha1") {
				ver.Algorithm = "SHA1"
				ver.CheckSum = string(utils.HTTPUtil{}.Get(a.GetBrowserDownloadURL()))[0:40] //Get first 40 character of SHA1 only
			}
		}

		// Older versions might have bad version patters ending up with a missing sha. Don't add them.
		if ver.CheckSum != "" {
			p.Versions = append(p.Versions, ver)
		}
	}
	return p
}
