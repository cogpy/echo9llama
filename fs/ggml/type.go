package ggml

import (
	"fmt"
	"log/slog"
	"strings"
)

// FileType is the Go equivalent to llama_ftype used for gguf file typing
type FileType uint32

const (
	FileTypeF32 FileType = iota
	FileTypeF16
	fileTypeQ4_0
	fileTypeQ4_1
	fileTypeMXFP4 // originally fileTypeQ4_1_F16 // unused by GGML
	fileTypeQ4_2  // unused by GGML
	fileTypeQ4_3  // unused by GGML
	FileTypeQ8_0
	fileTypeQ5_0
	fileTypeQ5_1
	fileTypeQ2K
	fileTypeQ3KS
	fileTypeQ3KM
	fileTypeQ3KL
	FileTypeQ4KS
	FileTypeQ4KM
	fileTypeQ5KS
	fileTypeQ5KM
	fileTypeQ6K
	fileTypeIQ2XXS
	fileTypeIQ2XS
	fileTypeQ2KS
	fileTypeIQ3XS
	fileTypeIQ3XXS
	fileTypeIQ1S
	fileTypeIQ4NL
	fileTypeIQ3S
	fileTypeIQ3M
	fileTypeIQ2S
	fileTypeIQ2M
	fileTypeIQ4XS
	fileTypeIQ1M
	FileTypeBF16
	fileTypeQ4_0_4_4 // unused by GGML
	fileTypeQ4_0_4_8 // unused by GGML
	fileTypeQ4_0_8_8 // unused by GGML
	fileTypeTQ1_0
	fileTypeTQ2_0

	FileTypeUnknown = 1024
)

// ParseFileType parses the provided GGUF file type
// Only Ollama supported types are considered valid
func ParseFileType(s string) (FileType, error) {
	switch s {
	case "F32":
		return FileTypeF32, nil
	case "F16":
		return FileTypeF16, nil
	case "Q8_0":
		return FileTypeQ8_0, nil
	case "Q4_K_S":
		return FileTypeQ4KS, nil
	case "Q4_K_M", "Q4_K":
		return FileTypeQ4KM, nil
	case "BF16":
		return FileTypeBF16, nil
	default:
		supportedFileTypes := []FileType{
			FileTypeF32,
			FileTypeF16,
			FileTypeQ4KS,
			FileTypeQ4KM,
			FileTypeQ8_0,
			// fsggml.FileTypeBF16, // TODO
		}
		strs := make([]string, len(supportedFileTypes))
		for i := range supportedFileTypes {
			strs[i] = supportedFileTypes[i].String()
		}

		return FileTypeUnknown, fmt.Errorf("unsupported quantization type %s - supported types are %s", s, strings.Join(strs, ", "))
	}
}

func (t FileType) String() string {
	// Note: this routine will return a broader set of file types for existing models
	switch t {
	case FileTypeF32:
		return "F32"
	case FileTypeF16:
		return "F16"
	case fileTypeQ4_0:
		return "Q4_0"
	case fileTypeQ4_1:
		return "Q4_1"
	case fileTypeMXFP4:
		return "MXFP4"
	case FileTypeQ8_0:
		return "Q8_0"
	case fileTypeQ5_0:
		return "Q5_0"
	case fileTypeQ5_1:
		return "Q5_1"
	case fileTypeQ2K:
		return "Q2_K"
	case fileTypeQ3KS:
		return "Q3_K_S"
	case fileTypeQ3KM:
		return "Q3_K_M"
	case fileTypeQ3KL:
		return "Q3_K_L"
	case FileTypeQ4KS:
		return "Q4_K_S"
	case FileTypeQ4KM:
		return "Q4_K_M"
	case fileTypeQ5KS:
		return "Q5_K_S"
	case fileTypeQ5KM:
		return "Q5_K_M"
	case fileTypeQ6K:
		return "Q6_K"
	case fileTypeQ2KS:
		return "Q2_K_S"
	case FileTypeBF16:
		return "BF16"
	default:
		return "unknown"
	}
}

func (t FileType) Value() uint32 {
	return uint32(t)
}

func (t FileType) ToTensorType() TensorType {
	switch t {
	case FileTypeF32:
		return TensorTypeF32
	case FileTypeF16:
		return TensorTypeF16
	case fileTypeQ4_0:
		return TensorTypeQ4_0
	case fileTypeQ4_1:
		return TensorTypeQ4_1
	case fileTypeMXFP4:
		return TensorTypeMXFP4 // Formerly unused tensorTypeQ4_2
	case FileTypeQ8_0:
		return TensorTypeQ8_0
	case fileTypeQ5_0:
		return TensorTypeQ5_0
	case fileTypeQ5_1:
		return TensorTypeQ5_1
	case fileTypeQ2K:
		return TensorTypeQ2K
	case fileTypeQ3KS:
		return TensorTypeQ3K
	case fileTypeQ3KM:
		return TensorTypeQ3K
	case fileTypeQ3KL:
		return TensorTypeQ3K
	case FileTypeQ4KS:
		return TensorTypeQ4K
	case FileTypeQ4KM:
		return TensorTypeQ4K
	case fileTypeQ5KS:
		return TensorTypeQ5K
	case fileTypeQ5KM:
		return TensorTypeQ5K
	case fileTypeQ6K:
		return TensorTypeQ6K
	case fileTypeQ2KS:
		return TensorTypeQ2K
	case FileTypeBF16:
		return TensorTypeBF16
	default:
		slog.Warn("unsupported file type", "type", t)
		return 0 // F32
	}
}

// TensorType is equivalent to ggml_type for individual tensor types
// Note: these are not the same as FileType
type TensorType uint32

const (
	TensorTypeF32 TensorType = iota
	TensorTypeF16
	TensorTypeQ4_0
	TensorTypeQ4_1
	TensorTypeMXFP4 // Formerly unused tensorTypeQ4_2
	tensorTypeQ4_3  // unused by GGML
	TensorTypeQ5_0
	TensorTypeQ5_1
	TensorTypeQ8_0
	TensorTypeQ8_1
	TensorTypeQ2K
	TensorTypeQ3K
	TensorTypeQ4K
	TensorTypeQ5K
	TensorTypeQ6K
	TensorTypeQ8K
	tensorTypeIQ2XXS // not supported by ollama
	tensorTypeIQ2XS  // not supported by ollama
	tensorTypeIQ3XXS // not supported by ollama
	tensorTypeIQ1S   // not supported by ollama
	tensorTypeIQ4NL  // not supported by ollama
	tensorTypeIQ3S   // not supported by ollama
	tensorTypeIQ2S   // not supported by ollama
	tensorTypeIQ4XS  // not supported by ollama
	TensorTypeI8
	TensorTypeI16
	TensorTypeI32
	TensorTypeI64
	TensorTypeF64
	tensorTypeIQ1M // not supported by ollama
	TensorTypeBF16
	tensorTypeQ4_0_4_4 // unused by GGML
	tensorTypeQ4_0_4_8 // unused by GGML
	tensorTypeQ4_0_8_8 // unused by GGML
	tensorTypeTQ1_0    // not supported by ollama
	tensorTypeTQ2_0    // not supported by ollama
	tensorTypeIQ4NL4x4 // unused by GGML
	tensorTypeIQ4NL4x8 // unused by GGML
	tensorTypeIQ4NL8x8 // unused by GGML
)

// ParseFileType parses the provided GGUF file type
// Only Ollama supported types are considered valid
func ParseTensorType(s string) (TensorType, error) {
	switch s {
	case "F32":
		return TensorTypeF32, nil
	case "F16":
		return TensorTypeF16, nil
	case "Q4_0":
		return TensorTypeQ4_0, nil
	case "Q4_1":
		return TensorTypeQ4_1, nil
	case "Q5_0":
		return TensorTypeQ5_0, nil
	case "Q5_1":
		return TensorTypeQ5_1, nil
	case "Q8_0":
		return TensorTypeQ8_0, nil
	case "Q8_1":
		return TensorTypeQ8_1, nil
	case "Q2_K":
		return TensorTypeQ2K, nil
	case "Q3_K":
		return TensorTypeQ3K, nil
	case "Q4_K":
		return TensorTypeQ4K, nil
	case "Q5_K":
		return TensorTypeQ5K, nil
	case "Q6_K":
		return TensorTypeQ6K, nil
	case "Q8_K":
		return TensorTypeQ8K, nil
	case "F64":
		return TensorTypeF64, nil
	case "BF16":
		return TensorTypeBF16, nil
	case "MXFP4":
		return TensorTypeMXFP4, nil
	default:
		return 0, fmt.Errorf("unsupported quantization type %s", s)
	}
}

func (t TensorType) IsQuantized() bool {
	switch t {
	case TensorTypeF32, TensorTypeF16, TensorTypeBF16:
		return false
	default:
		return true
	}
}

func (t TensorType) RowSize(ne uint64) uint64 {
	return t.TypeSize() * ne / t.BlockSize()
}

func (t TensorType) String() string {
	switch t {
	case TensorTypeF32:
		return "F32"
	case TensorTypeF16:
		return "F16"
	case TensorTypeQ4_0:
		return "Q4_0"
	case TensorTypeQ4_1:
		return "Q4_1"
	case TensorTypeQ5_0:
		return "Q5_0"
	case TensorTypeQ5_1:
		return "Q5_1"
	case TensorTypeQ8_0:
		return "Q8_0"
	case TensorTypeQ8_1:
		return "Q8_1"
	case TensorTypeQ2K:
		return "Q2_K"
	case TensorTypeQ3K:
		return "Q3_K"
	case TensorTypeQ4K:
		return "Q4_K"
	case TensorTypeQ5K:
		return "Q5_K"
	case TensorTypeQ6K:
		return "Q6_K"
	case TensorTypeQ8K:
		return "Q8_K"
	case TensorTypeF64:
		return "F64"
	case TensorTypeBF16:
		return "BF16"
	case TensorTypeMXFP4:
		return "MXFP4"
	default:
		return "unknown"
	}
}
