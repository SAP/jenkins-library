package buildpacks

type PathEnum string

const (
	PathEnumRoot    = PathEnum("root")
	PathEnumFolder  = PathEnum("folder")
	PathEnumArchive = PathEnum("archive")
)
