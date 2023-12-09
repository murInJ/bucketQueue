package core

import "sync"

type tagController struct {
	tagMap sync.Map
}

func (tc *tagController) ClaimTag(tag any) bool {
	_, loaded := tc.tagMap.LoadOrStore(tag, true)
	return !loaded
}

func (tc *tagController) ReleaseTag(tag any) {
	tc.tagMap.Delete(tag)
}
