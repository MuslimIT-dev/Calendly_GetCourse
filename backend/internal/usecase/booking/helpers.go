package booking

import (
	"context"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

func isMasterOwner(ctx context.Context, masters domain.MasterRepository, userID, appointmentMasterID int32) bool {
	if appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return true
	}
	master, err := masters.GetByUserID(ctx, userID)
	if err != nil {
		return false
	}
	return master.ID == appointmentMasterID
}

func isMasterOrClient(ctx context.Context, masters domain.MasterRepository, userID, appointmentMasterID, appointmentClientID int32) bool {
	if userID == appointmentClientID {
		return true
	}
	return isMasterOwner(ctx, masters, userID, appointmentMasterID)
}