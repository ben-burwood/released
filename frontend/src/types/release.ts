export interface Release {
  tag: string;
  name: string;
  bodyHtml?: string;
  publishedAt: string | null;
  author?: string;
  prerelease: boolean;
  draft: boolean;
  url: string;
}

export interface Releases {
  repo: string;
  releases: Release[];
  fetchedAt: string;
}
