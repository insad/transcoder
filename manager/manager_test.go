package manager

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/lbryio/transcoder/pkg/logging"
	"github.com/lbryio/transcoder/pkg/mfr"
	"github.com/lbryio/transcoder/storage"
	"github.com/lbryio/transcoder/video"
	"github.com/stretchr/testify/suite"
)

type managerSuite struct {
	suite.Suite
}

func isLevel5(key string) bool {
	return rand.Intn(2) == 0
}

func isChannelEnabled(key string) bool {
	return rand.Intn(2) == 0
}

func TestManagerSuite(t *testing.T) {
	suite.Run(t, new(managerSuite))
}

type vlib struct {
	ret *video.Video
}

func (l *vlib) Get(h string) (*video.Video, error) {
	return l.ret, nil
}

func (l *vlib) Add(_ video.AddParams) (*video.Video, error) {
	return nil, nil
}

func (l *vlib) New(_ string) *storage.LocalStream {
	return &storage.LocalStream{}
}

func (l *vlib) Path() string {
	return ""
}

func (l *vlib) AddLocalStream(_, _ string, _ storage.LocalStream) (*video.Video, error) {
	return nil, nil
}

func (l *vlib) AddRemoteStream(storage.RemoteStream) (*video.Video, error) {
	return nil, nil
}

func (s *managerSuite) SetupSuite() {
	logger = logging.Create("manager", logging.Dev)
}

func (s *managerSuite) TestVideo() {
	mgr := NewManager(&vlib{ret: nil}, 0)

	LoadConfiguredChannels(
		[]string{
			"@BretWeinstein:f",
		},
		[]string{
			"@davidpakman#7",
			"@specialoperationstest#3",
		},
		[]string{
			"@TheVoiceofReason#a",
		},
	)

	urlsPriority := []string{
		"@BretWeinstein#f/EvoLens87#1",
	}
	urlsEnabled := []string{
		"@davidpakman#7/vaccination-delays-and-more-biden-picks#8",
		"@specialoperationstest#3/fear-of-death-inspirational#a",
	}
	urlsLevel5 := []string{
		"@samtime#1/airpods-max-parody-ehh-pods-max#7",
	}
	urlsNotEnabled := []string{
		"@TRUTH#2/what-do-you-know-what-do-you-believe#2",
	}
	urlsNoChannel := []string{
		"what#1",
	}
	urlsDisabled := []string{
		"lbry://@TheVoiceofReason#a/PaypalSucks#5",
	}
	urlsNotFound := []string{
		randomString(96),
		randomString(24) + "#" + randomString(12),
		randomString(500),
	}

	for _, u := range urlsPriority {
		v, err := mgr.Video(u)
		s.Nil(v)
		s.Equal(ErrTranscodingQueued, err)
	}

	for _, u := range urlsEnabled {
		v, err := mgr.Video(u)
		s.Nil(v)
		s.Equal(ErrTranscodingQueued, err)
	}

	for _, u := range urlsLevel5 {
		v, err := mgr.Video(u)
		s.Nil(v)
		s.Equal(ErrTranscodingQueued, err)
	}

	for _, u := range urlsNotEnabled {
		v, err := mgr.Video(u)
		s.Nil(v)
		s.Equal(ErrTranscodingForbidden, err)
	}

	for _, u := range urlsDisabled {
		v, err := mgr.Video(u)
		s.Nil(v)
		s.Equal(ErrChannelNotEnabled, err)
	}

	for _, u := range urlsNoChannel {
		v, err := mgr.Video(u)
		s.Nil(v)
		s.Equal(ErrNoSigningChannel, err)
	}

	for _, u := range urlsNotFound {
		v, err := mgr.Video(u)
		s.Nil(v)
		s.Equal(ErrStreamNotFound, err)
	}

	expectedUrls := []string{urlsPriority[0], urlsEnabled[0], urlsLevel5[0], urlsNotEnabled[0], urlsEnabled[1]}
	receivedUrls := []string{}
	for r := range mgr.Requests() {
		receivedUrls = append(receivedUrls, r.URI)
		if len(receivedUrls) == len(expectedUrls) {
			mgr.pool.Stop()
			break
		}
	}
	sort.Strings(expectedUrls)
	sort.Strings(receivedUrls)
	s.Equal(expectedUrls, receivedUrls)

}

func (s *managerSuite) TestRequests() {
	var r1, r2 *TranscodingRequest

	LoadConfiguredChannels(
		[]string{},
		[]string{
			"@specialoperationstest#3",
		},
		[]string{},
	)

	mgr := NewManager(&vlib{ret: nil}, 0)
	mgr.Video("@specialoperationstest#3/fear-of-death-inspirational#a")
	out := mgr.Requests()
	r1 = <-out

	s.Equal(mfr.StatusActive, mgr.RequestStatus(r1.SDHash))
	select {
	case r2 = <-out:
		s.Failf("got output from Requests channel", "%v", r2)
	default:
	}

	s.NotNil(r1)
}

func TestValidateIncomingVideo(t *testing.T) {
}

func randomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}
