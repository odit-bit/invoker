package crawler

import (
	"context"
	"testing"

	"github.com/google/uuid"
	mock_crawler "github.com/odit-bit/invoker/linkcrawler/mocks"
	"go.uber.org/mock/gomock"
)

func Test_graph_updater(t *testing.T) {
	ctrl := gomock.NewController(t)
	gu := mock_crawler.NewMockGraphUpdater(ctrl)

	gu.EXPECT().UpsertLink(gomock.Any()).AnyTimes().
		Return(nil)

	gu.EXPECT().UpsertEdge(gomock.Any()).AnyTimes().
		Return(nil)

	gu.EXPECT().RemoveStaleEdges(gomock.Any(), gomock.Any()).AnyTimes().
		Return(nil)

	updater := newUpdater(gu)

	p := payload{
		LinkID: uuid.New(),
		URL:    "http://source.com",
		NoFollowLinks: []string{
			"http://noFollow.com",
		},
		Links: []string{
			"http://follow_1.com/foo",
			"http://follow_2.com/bar",
		},
	}

	result, err := updater.Process(context.TODO(), &p)
	if err != nil {
		t.Error(err)
	}
	_ = result
}

func Test_graph_updater_integration(t *testing.T) {

}
