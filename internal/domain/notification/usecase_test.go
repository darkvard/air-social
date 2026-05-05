package notification_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	chatmocks "air-social/internal/domain/chat/mocks"
	"air-social/internal/domain/common"
	notif "air-social/internal/domain/notification"
	notifmocks "air-social/internal/domain/notification/mocks"
	"air-social/pkg"
)

type NotificationUseCaseSuite struct {
	suite.Suite
}

func TestNotificationUseCaseSuite(t *testing.T) {
	suite.Run(t, new(NotificationUseCaseSuite))
}

type testDeps struct {
	repo     *notifmocks.MockRepository
	presence *chatmocks.MockPresenceStore
	pusher   *notifmocks.MockNotificationPusher
}

func (s *NotificationUseCaseSuite) newUC(deps testDeps) notif.UseCase {
	return notif.NewUseCase(notif.Deps{
		Repo:     deps.repo,
		Presence: deps.presence,
		Pusher:   deps.pusher,
	})
}

func (s *NotificationUseCaseSuite) makeDeps() testDeps {
	return testDeps{
		repo:     notifmocks.NewMockRepository(s.T()),
		presence: chatmocks.NewMockPresenceStore(s.T()),
		pusher:   notifmocks.NewMockNotificationPusher(s.T()),
	}
}

// TestCreate

func (s *NotificationUseCaseSuite) TestCreate() {
	userID := int64(10)
	actorID := int64(20)
	targetID := int64(99)

	n := &notif.Notification{
		UserID:  userID,
		ActorID: actorID,
		Type:    notif.NotifPostLiked,
		TargetID: &targetID,
	}

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "repo_error_returns_error",
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					Upsert(mock.Anything, n).
					Return(pkg.ErrInternal).
					Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "user_offline_no_push",
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					Upsert(mock.Anything, n).
					Return(nil).
					Once()
				deps.presence.EXPECT().
					IsOnlineBatch(mock.Anything, []int64{userID}).
					Return(map[int64]bool{userID: false}, nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "user_online_push_succeeds",
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					Upsert(mock.Anything, n).
					Return(nil).
					Once()
				deps.presence.EXPECT().
					IsOnlineBatch(mock.Anything, []int64{userID}).
					Return(map[int64]bool{userID: true}, nil).
					Once()
				deps.pusher.EXPECT().
					PushNotification(mock.Anything, userID, n).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "presence_error_skips_push_no_error",
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					Upsert(mock.Anything, n).
					Return(nil).
					Once()
				deps.presence.EXPECT().
					IsOnlineBatch(mock.Anything, []int64{userID}).
					Return(nil, pkg.ErrInternal).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "push_error_is_swallowed",
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					Upsert(mock.Anything, n).
					Return(nil).
					Once()
				deps.presence.EXPECT().
					IsOnlineBatch(mock.Anything, []int64{userID}).
					Return(map[int64]bool{userID: true}, nil).
					Once()
				deps.pusher.EXPECT().
					PushNotification(mock.Anything, userID, n).
					Return(pkg.ErrInternal).
					Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			deps := s.makeDeps()
			tc.setupMock(deps)
			err := s.newUC(deps).Create(context.Background(), n)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

// TestDelete

func (s *NotificationUseCaseSuite) TestDelete() {
	userID := int64(10)
	actorID := int64(20)
	targetID := int64(99)

	tests := []struct {
		name      string
		targetID  *int64
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name:     "with_target_id_success",
			targetID: &targetID,
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					Delete(mock.Anything, userID, actorID, notif.NotifPostLiked, &targetID).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name:     "without_target_id_success",
			targetID: nil,
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					Delete(mock.Anything, userID, actorID, notif.NotifUserFollowed, (*int64)(nil)).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name:     "repo_error_propagates",
			targetID: &targetID,
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					Delete(mock.Anything, userID, actorID, notif.NotifPostLiked, &targetID).
					Return(pkg.ErrNotFound).
					Once()
			},
			wantErr: pkg.ErrNotFound,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			deps := s.makeDeps()
			tc.setupMock(deps)
			typ := notif.NotifPostLiked
			if tc.targetID == nil {
				typ = notif.NotifUserFollowed
			}
			err := s.newUC(deps).Delete(context.Background(), userID, actorID, typ, tc.targetID)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

// TestGetNotifications

func (s *NotificationUseCaseSuite) TestGetNotifications() {
	userID := int64(10)

	items := []notif.Notification{
		{ID: 5, UserID: userID, Type: notif.NotifPostLiked},
		{ID: 4, UserID: userID, Type: notif.NotifCommentLiked},
	}

	tests := []struct {
		name      string
		params    notif.GetParams
		setupMock func(deps testDeps)
		wantLen   int
		wantErr   error
	}{
		{
			name:   "success_returns_items",
			params: notif.GetParams{UserID: userID, Query: common.CursorQueryParams[int64]{Limit: 20}},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetList(mock.Anything, mock.AnythingOfType("notification.GetParams")).
					Return(items, nil).
					Once()
			},
			wantLen: 2,
			wantErr: nil,
		},
		{
			name:   "repo_error",
			params: notif.GetParams{UserID: userID, Query: common.CursorQueryParams[int64]{Limit: 20}},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetList(mock.Anything, mock.AnythingOfType("notification.GetParams")).
					Return(nil, pkg.ErrInternal).
					Once()
			},
			wantLen: 0,
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "empty_result",
			params: notif.GetParams{UserID: userID, Query: common.CursorQueryParams[int64]{Limit: 20}},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetList(mock.Anything, mock.AnythingOfType("notification.GetParams")).
					Return([]notif.Notification{}, nil).
					Once()
			},
			wantLen: 0,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			deps := s.makeDeps()
			tc.setupMock(deps)
			result, err := s.newUC(deps).GetNotifications(context.Background(), tc.params)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Empty(result.Data)
			} else {
				s.NoError(err)
				s.Len(result.Data, tc.wantLen)
			}
		})
	}
}

// TestGetUnreadCount

func (s *NotificationUseCaseSuite) TestGetUnreadCount() {
	userID := int64(10)

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantCount int64
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetUnreadCount(mock.Anything, userID).
					Return(int64(3), nil).
					Once()
			},
			wantCount: 3,
			wantErr:   nil,
		},
		{
			name: "repo_error",
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetUnreadCount(mock.Anything, userID).
					Return(int64(0), pkg.ErrInternal).
					Once()
			},
			wantCount: 0,
			wantErr:   pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			deps := s.makeDeps()
			tc.setupMock(deps)
			count, err := s.newUC(deps).GetUnreadCount(context.Background(), userID)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
				s.Equal(tc.wantCount, count)
			}
		})
	}
}

// TestMarkRead

func (s *NotificationUseCaseSuite) TestMarkRead() {
	userID := int64(10)
	notifID := int64(42)

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					MarkRead(mock.Anything, userID, notifID).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "not_found",
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					MarkRead(mock.Anything, userID, notifID).
					Return(pkg.ErrNotFound).
					Once()
			},
			wantErr: pkg.ErrNotFound,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			deps := s.makeDeps()
			tc.setupMock(deps)
			err := s.newUC(deps).MarkRead(context.Background(), userID, notifID)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

// TestMarkAllRead

func (s *NotificationUseCaseSuite) TestMarkAllRead() {
	userID := int64(10)

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					MarkAllRead(mock.Anything, userID).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "repo_error",
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					MarkAllRead(mock.Anything, userID).
					Return(pkg.ErrInternal).
					Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			deps := s.makeDeps()
			tc.setupMock(deps)
			err := s.newUC(deps).MarkAllRead(context.Background(), userID)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}
