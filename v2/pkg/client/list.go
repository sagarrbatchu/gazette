package client

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	pb "github.com/LiveRamp/gazette/v2/pkg/protocol"
	log "github.com/sirupsen/logrus"
)

// PolledList performs periodic polls of a ListRequest. Its most recent
// polled result may be accessed via List.
type PolledList struct {
	ctx    context.Context
	client pb.JournalClient
	req    pb.ListRequest
	resp   atomic.Value
}

// NewPolledList returns a PolledList of the ListRequest which is initialized and
// ready for immediate use, and which will regularly refresh with interval |dur|.
// An error encountered in the first List RPC is returned. Subsequent RPC errors
// will be logged as warnings and retried as part of regular refreshes.
func NewPolledList(ctx context.Context, client pb.JournalClient, dur time.Duration, req pb.ListRequest) (*PolledList, error) {
	var resp, err = ListAll(ctx, client, req)
	if err != nil {
		return nil, err
	}
	var pl = &PolledList{ctx: ctx, client: client, req: req}
	pl.resp.Store(resp)

	go pl.periodicRefresh(dur)
	return pl, nil
}

// List returns the most recent ListResponse.
func (pl *PolledList) List() *pb.ListResponse { return pl.resp.Load().(*pb.ListResponse) }

func (pl *PolledList) periodicRefresh(dur time.Duration) {
	var ticker = time.NewTicker(dur)
	for {
		select {
		case <-ticker.C:
			var resp, err = ListAll(pl.ctx, pl.client, pl.req)
			if err != nil {
				log.WithFields(log.Fields{"err": err, "req": pl.req.String()}).Warn("periodic List refresh failed")
			} else {
				pl.resp.Store(resp)
			}
		case <-pl.ctx.Done():
			ticker.Stop()
			return
		}
	}
}

// ListAll performs multiple List RPCs, as required to join across multiple
// ListResponse pages, and returns the complete ListResponse of the ListRequest.
// Any encountered error is returned.
func ListAll(ctx context.Context, client pb.JournalClient, req pb.ListRequest) (*pb.ListResponse, error) {
	var resp *pb.ListResponse

	for {
		// List RPCs may be dispatched to any broker.
		if r, err := client.List(pb.WithDispatchDefault(ctx), &req); err != nil {
			return resp, err
		} else if err = r.Validate(); err != nil {
			return resp, err
		} else if r.Status != pb.Status_OK {
			return resp, errors.New(r.Status.String())
		} else {
			req.PageToken, r.NextPageToken = r.NextPageToken, ""

			if resp == nil {
				resp = r
			} else {
				resp.Journals = append(resp.Journals, r.Journals...)
			}
		}
		if req.PageToken == "" {
			break // All done.
		}
	}

	if dr, ok := client.(pb.DispatchRouter); ok {
		for _, j := range resp.Journals {
			dr.UpdateRoute(j.Spec.Name.String(), &j.Route)
		}
	}
	return resp, nil
}

// ApplyJournals invokes the Apply RPC, and maps a validation or !OK status to an error.
func ApplyJournals(ctx context.Context, jc pb.JournalClient, req *pb.ApplyRequest) (*pb.ApplyResponse, error) {
	if r, err := jc.Apply(pb.WithDispatchDefault(ctx), req); err != nil {
		return r, err
	} else if err = r.Validate(); err != nil {
		return r, err
	} else if r.Status != pb.Status_OK {
		return r, errors.New(r.Status.String())
	} else {
		return r, nil
	}
}
