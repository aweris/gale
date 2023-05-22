package model

// Repository represents a GitHub repository
type Repository struct {
	ID               string
	Name             string
	NameWithOwner    string
	URL              string
	Owner            RepositoryOwner
	DefaultBranchRef RepositoryBranchRef
}

// RepositoryOwner represents a GitHub repository owner
type RepositoryOwner struct {
	ID    string
	Login string
}

// RepositoryBranchRef represents a GitHub repository branch ref
type RepositoryBranchRef struct {
	Name string
}
