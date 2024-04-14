package bot

import "time"

type UserData struct {
	UserID         int64
	QueriesToday   int
	LastQueryTime  time.Time
	PendingQueires int
}

func (h *Handler) isAdmin(userID int64) bool {
	return userID == h.AdminID
}

func (h *Handler) adminRestricted() bool {
	return h.AdminID != 0 && h.RestrictToAdminID
}

// create a new user data if not exists
func (h *Handler) initUser(userID int64) {
	h.usersMapLock.Lock()
	defer h.usersMapLock.Unlock()

	if _, ok := h.usersMap[userID]; !ok {
		h.usersMap[userID] = &UserData{
			UserID: userID,
		}
	}
}

func (h *Handler) incrPending(userID int64) {
	h.usersMapLock.Lock()
	defer h.usersMapLock.Unlock()
	h.usersMap[userID].PendingQueires++
}

func (h *Handler) decrPending(userID int64) {
	h.usersMapLock.Lock()
	defer h.usersMapLock.Unlock()
	h.usersMap[userID].PendingQueires--
}

func (h *Handler) incrQueries(userID int64) {
	h.usersMapLock.Lock()
	defer h.usersMapLock.Unlock()
	h.usersMap[userID].QueriesToday++
	h.usersMap[userID].LastQueryTime = h.NowFunc()
}

type reason string

const (
	reasonNoUser  reason = "no user"
	reasonLimit   reason = "limit"
	reasonPending reason = "pending"
)

// check if QueriesToday is need to be reset
func (h *Handler) updateQueryCountLimit(userID int64) {
	h.usersMapLock.Lock()
	defer h.usersMapLock.Unlock()

	data, ok := h.usersMap[userID]
	if !ok {
		return
	}

	if data.LastQueryTime.Day() != h.NowFunc().Day() {
		data.QueriesToday = 0
	}
}

// admin can query without limits
func (h *Handler) canQuery(userID int64) (bool, reason) {
	h.usersMapLock.RLock()
	defer h.usersMapLock.RUnlock()

	if h.isAdmin(userID) {
		return true, ""
	}

	data, ok := h.usersMap[userID]
	if !ok {
		return false, reasonNoUser
	}

	if data.QueriesToday >= h.LimitPerDay {
		return false, reasonLimit
	}

	if data.PendingQueires >= h.LimitPending {
		return false, reasonPending
	}

	return true, ""
}
