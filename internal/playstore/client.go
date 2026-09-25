package playstore

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/playdeveloperreporting/v1beta1"
)

type Client struct {
	publisher *androidpublisher.Service
	reporting *playdeveloperreporting.Service
}

func NewClient(ctx context.Context, serviceAccountJSON []byte) (*Client, error) {
	creds, err := google.CredentialsFromJSONWithParams(ctx, serviceAccountJSON, google.CredentialsParams{
		Scopes: []string{
			androidpublisher.AndroidpublisherScope,
			playdeveloperreporting.PlaydeveloperreportingScope,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("parse service account credentials: %w", err)
	}

	publisher, err := androidpublisher.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("create android publisher client: %w", err)
	}

	reporting, err := playdeveloperreporting.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("create play developer reporting client: %w", err)
	}

	return &Client{publisher: publisher, reporting: reporting}, nil
}

type ReleaseInput struct {
	PackageName  string
	Track        string
	VersionCode  int64
	ReleaseNotes string
	BundlePath   string
}

func (c *Client) PublishBundle(ctx context.Context, in ReleaseInput) error {
	edit, err := c.publisher.Edits.Insert(in.PackageName, &androidpublisher.AppEdit{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("create edit: %w", err)
	}

	bundleFile, err := openFile(in.BundlePath)
	if err != nil {
		return err
	}
	defer bundleFile.Close()

	uploadedBundle, err := c.publisher.Edits.Bundles.Upload(in.PackageName, edit.Id).
		Media(bundleFile).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("upload bundle: %w", err)
	}

	release := &androidpublisher.TrackRelease{
		Name:         in.ReleaseNotes,
		Status:       "completed",
		VersionCodes: []int64{uploadedBundle.VersionCode},
	}
	if in.ReleaseNotes != "" {
		release.ReleaseNotes = []*androidpublisher.LocalizedText{
			{Language: "en-US", Text: in.ReleaseNotes},
		}
	}

	_, err = c.publisher.Edits.Tracks.Update(in.PackageName, edit.Id, in.Track, &androidpublisher.Track{
		Track:    in.Track,
		Releases: []*androidpublisher.TrackRelease{release},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("assign track: %w", err)
	}

	if _, err := c.publisher.Edits.Commit(in.PackageName, edit.Id).Context(ctx).Do(); err != nil {
		return fmt.Errorf("commit edit: %w", err)
	}

	return nil
}
