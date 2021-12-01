package imageConversion

type ConversionRequest struct {
	CompressTo *string // string representation of a file type we want to encode to, e.g. jpeg
	Suffix     string  // The name to append at the end of the file

	LongestSide *uint  // Longest side in the resize operation
	Obfuscate   *bool  // Whether to keep all filenames the same or randomize the names
	ResizeOp    string // string representation of what kind of resize operation will take place
}
