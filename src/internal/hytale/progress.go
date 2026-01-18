package hytale

// ProgressCallback is called with download progress (0.0 to 1.0)
type ProgressCallback func(percent float64, label string)
