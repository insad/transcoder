package testing

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"testing"
	"time"

	"github.com/lbryio/transcoder/db"
	"github.com/lbryio/transcoder/manager"
	"github.com/lbryio/transcoder/pkg/mfr"
	"github.com/lbryio/transcoder/storage"
	"github.com/lbryio/transcoder/video"
	"github.com/lbryio/transcoder/workers"
	"github.com/stretchr/testify/suite"
)

type integSuite struct {
	suite.Suite
	db      *db.DB
	lib     *video.Library
	mgr     *manager.VideoManager
	httpAPI *manager.HttpAPI

	assetsPath string
}

func TestWorkersSuite(t *testing.T) {
	suite.Run(t, new(integSuite))
}

func (s *integSuite) SetupSuite() {
	assetsPath := path.Join(s.assetsPath, "videos")
	s.db = db.OpenTestDB()
	s.Require().NoError(s.db.MigrateUp(video.InitialMigration))

	libCfg := video.Configure().
		LocalStorage(storage.Local(assetsPath)).
		DB(s.db)

	s.lib = video.NewLibrary(libCfg)
	s.mgr = manager.NewManager(s.lib, 10)

	workers.SpawnEncoderWorkers(3, s.mgr)
	s.httpAPI = manager.NewHttpAPI(
		manager.ConfigureHttpAPI().
			Debug(true).
			Addr("127.0.0.1:58018").
			VideoPath(assetsPath).
			VideoManager(s.mgr),
	)

	go func() {
		err := s.httpAPI.Start()
		if err != nil {
			s.FailNow(err.Error())
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	wait := time.NewTicker(50 * time.Millisecond)
	defer cancel()

	func() {
		for {
			select {
			case <-ctx.Done():
				s.FailNow("server startup taking too long")
			case <-wait.C:
				_, err := http.Get(fmt.Sprintf("http://%v/", s.httpAPI.Addr()))
				if err == nil {
					return
				}
			}
		}
	}()
}

func (s *integSuite) TestStreamNotFound() {
	resp, err := http.Get(fmt.Sprintf("http://%v/api/v2/video/%v", s.httpAPI.Addr(), randomString(25)))
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *integSuite) TestStreamQueuedLevel5() {
	lbryUrl := "@FreeMovies#a/the-jack-knife-man#f"
	escUrl := url.PathEscape("@FreeMovies#a/the-jack-knife-man#f")
	tr, err := manager.ResolveRequest(lbryUrl)
	s.Require().NoError(err)

	resp, err := http.Get(fmt.Sprintf("http://%v/api/v2/video/%v", s.httpAPI.Addr(), escUrl))
	s.Require().NoError(err)
	s.equalResponse(http.StatusAccepted, manager.ErrTranscodingQueued.Error(), resp)

	time.Sleep(1 * time.Second)

	s.Equal(mfr.StatusActive, s.mgr.RequestStatus(tr.SDHash))

	resp, err = http.Get(fmt.Sprintf("http://%v/api/v2/video/%v", s.httpAPI.Addr(), escUrl))

	s.Equal(mfr.StatusActive, s.mgr.RequestStatus(tr.SDHash))

	s.Require().NoError(err)
	s.equalResponse(http.StatusAccepted, manager.ErrTranscodingUnderway.Error(), resp)
}

// func (s *integSuite) TestStreamQueuedEnabled() {
// 	lbryUrl := url.PathEscape("@FreeMovies#a/the-jack-knife-man#f")
// 	resp, err := http.Get(
// 		fmt.Sprintf("http://%v/api/v2/video/%v", s.httpAPI.Addr(), lbryUrl))
// 	s.Require().NoError(err)
// 	s.equalResponse(http.StatusAccepted, "", resp)
// 	time.Sleep(10 * time.Second)
// }

func (s *integSuite) TestStreamQueuedCommon() {
	lbryUrl := "@specialoperationstest#3/fear-of-death-inspirational#a"
	escUrl := url.PathEscape("@specialoperationstest#3/fear-of-death-inspirational#a")
	tr, err := manager.ResolveRequest(lbryUrl)
	s.Require().NoError(err)

	resp, err := http.Get(fmt.Sprintf("http://%v/api/v2/video/%v", s.httpAPI.Addr(), escUrl))
	s.Require().NoError(err)
	s.equalResponse(http.StatusForbidden, manager.ErrTranscodingForbidden.Error(), resp)
	s.Equal(mfr.StatusQueued, s.mgr.RequestStatus(tr.SDHash))
}

func (s *integSuite) equalResponse(expCode int, expBody string, resp *http.Response) {
	s.Require().Equal(expCode, resp.StatusCode)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.FailNowf("cannot read response body: %v", err.Error())
	}
	s.Equal(expBody, string(body))
}

func randomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}
