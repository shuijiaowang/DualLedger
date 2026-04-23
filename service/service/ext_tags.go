package service

import "SService/model"

const extTagsKey = "tags"

func tagsFromExt(ext model.JSONMap) []string {
	if ext == nil {
		return nil
	}
	raw, ok := ext[extTagsKey]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return append([]string(nil), v...)
	case model.JSONStrings:
		return append([]string(nil), v...)
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if ok && s != "" {
				out = append(out, s)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return nil
	}
}
