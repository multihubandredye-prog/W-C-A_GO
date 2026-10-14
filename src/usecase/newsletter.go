package usecase

import (
	"context"

	domainNewsletter "github.com/multihubandredye-prog/W-C-A_GO/domains/newsletter"
	"github.com/multihubandredye-prog/W-C-A_GO/infrastructure/whatsapp"
	pkgError "github.com/multihubandredye-prog/W-C-A_GO/pkg/error"
	"github.com/multihubandredye-prog/W-C-A_GO/pkg/utils"
	"github.com/multihubandredye-prog/W-C-A_GO/validations"
)

type serviceNewsletter struct{}

func NewNewsletterService() domainNewsletter.INewsletterUsecase {
	return &serviceNewsletter{}
}

func (service serviceNewsletter) Unfollow(ctx context.Context, request domainNewsletter.UnfollowRequest) (err error) {
	if err = validations.ValidateUnfollowNewsletter(ctx, request); err != nil {
		return err
	}

	client := whatsapp.ClientFromContext(ctx)
	if client == nil {
		return pkgError.ErrWaCLI
	}

	JID, err := utils.ValidateJidWithLogin(client, request.NewsletterID)
	if err != nil {
		return err
	}

	return client.UnfollowNewsletter(ctx, JID)
}
