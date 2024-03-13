package bot_server

import (
	"context"
	"rock_review/util/goutil"
	"rock_review/util/xlogger"
	"sync"
	"time"
)

type UserSessionMgr struct {
	repo *UserSessionRepo

	inactiveSeconds  int64
	m                sync.Mutex
	activeSessionMap map[int64]*UserSession
}

func NewUserSessionMgr(repo *UserSessionRepo) *UserSessionMgr {
	mgr := &UserSessionMgr{
		repo:             repo,
		inactiveSeconds:  300,
		m:                sync.Mutex{},
		activeSessionMap: map[int64]*UserSession{},
	}

	goutil.SafeGo(context.TODO(), func() {
		mgr.RunRoutine(context.TODO())
	})

	return mgr
}

func (mgr *UserSessionMgr) GetCurrentUserSession(ctx context.Context, userId int64, chatId int64, botSvc *ReviewBotSvc) *UserSession {
	var (
		now = time.Now()
	)

	mgr.m.Lock()
	userSession, ok := mgr.activeSessionMap[userId]
	if userSession != nil {
		userSession.lastActiveUnix = now.Unix()
	}
	mgr.m.Unlock()
	if ok {
		return userSession
	}

	// load session data to create new session
	sessionData, err := mgr.repo.GetUserSessionData(ctx, userId)
	if err != nil {
		xlogger.ErrorF(ctx, "get user session data fail: %v", err)
		sessionData = mgr.repo.newInitSessionData(userId)
	}

	// init session
	userSession = NewUserSession(sessionData, chatId, botSvc, botSvc.reviewRepo, mgr.repo)

	// register session if not exist
	mgr.m.Lock()
	existSession, ok := mgr.activeSessionMap[userId]
	if ok {
		userSession = existSession
	} else {
		mgr.activeSessionMap[userId] = userSession
	}
	mgr.m.Unlock()

	// run session
	goutil.SafeGo(ctx, func() {
		userSession.Run(ctx)
	})

	return userSession
}

func (mgr *UserSessionMgr) RunRoutine(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mgr.purgeInactiveSession(ctx)
		}
	}
}

func (mgr *UserSessionMgr) purgeInactiveSession(ctx context.Context) {
	now := time.Now()
	inactiveDeadline := now.Unix() - mgr.inactiveSeconds

	mgr.m.Lock()
	for userId, session := range mgr.activeSessionMap {
		mgr.m.Unlock()
		purged := false

		mgr.m.Lock()
		if session.lastActiveUnix < inactiveDeadline {
			delete(mgr.activeSessionMap, userId)
			purged = true
		}
		mgr.m.Unlock()

		if purged {
			session.Close()
			xlogger.InfoF(ctx, "session purged due to inactive, user: %d", userId)
		}

		mgr.m.Lock()
	}
	mgr.m.Unlock()
}
