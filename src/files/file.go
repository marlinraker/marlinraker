package files

type FileMeta struct {
	Modified    float64 `json:"modified"`
	Size        int64   `json:"size"`
	Permissions string  `json:"permissions"`
	FileName    string  `json:"filename"`
}
