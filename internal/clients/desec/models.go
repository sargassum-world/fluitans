package desec

import (
	"time"

	"github.com/sargassum-world/fluitans/pkg/desec"
)

type DesecAPISettings struct {
	ReadCacheTTL   time.Duration
	WriteSoftQuota float32
}

type RRsetKey struct {
	Subname string
	Type    string
}

func NewRRsetKey(rrset desec.RRset) RRsetKey {
	return RRsetKey{
		Subname: rrset.Subname,
		Type:    rrset.Type,
	}
}

func (k RRsetKey) AsDeletionUpsertRRset() desec.RRset {
	return desec.RRset{
		Subname: k.Subname,
		Type:    k.Type,
		Records: []string{},
	}
}

func IsDeletionUpsertRRset(rrset desec.RRset) bool {
	return rrset.Records != nil && len(rrset.Records) == 0
}
