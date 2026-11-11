package relevance

import "reflect"

// copyExportedFields copies every exported field of src into dst. Both must be
// pointers to the same struct type. Unexported fields such as the embedded
// mutex are left untouched, so callers receive a snapshot without a lock copy.
func copyExportedFields(dst, src interface{}) {
	d := reflect.ValueOf(dst).Elem()
	s := reflect.ValueOf(src).Elem()
	for i := 0; i < s.NumField(); i++ {
		if s.Type().Field(i).PkgPath != "" {
			continue // unexported
		}
		d.Field(i).Set(s.Field(i))
	}
}
