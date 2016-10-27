package member

import (
	"github.com/dropbox/dropbox-sdk-go-unofficial/team"
	"github.com/watermint/dreport/auth"
	"github.com/watermint/dreport/crawler"
	"github.com/watermint/dreport/integration"
	"strconv"
)

type ReportMemberProfile struct {
}

func (t *ReportMemberProfile) ReportName() string {
	return "TeamMemberProfile"
}

func (t *ReportMemberProfile) ReportDescription() string {
	return "List all team member profiles of a team"
}

func (t *ReportMemberProfile) RequiredPermissions() []string {
	return []string{auth.PERMISSION_INFO}
}

func (t *ReportMemberProfile) Report(context *integration.ReportContext) error {
	members, err := crawler.AllTeamMembers(context)
	if err != nil {
		return err
	}

	context.ReportOutput.Headers(t.createHeader())

	for _, m := range members {
		context.ReportOutput.Row(t.createRow(m))
	}

	return nil
}

func (t *ReportMemberProfile) createHeader() []string {
	return []string{
		"Account Id",
		"Team Member Id",
		"Email",
		"Email verified?",
		"External Id",
		"Membership Type",
		"Role",
		"Status",
	}
}

func (t *ReportMemberProfile) createRow(member *team.TeamMemberInfo) []string {
	return []string{
		member.Profile.AccountId,
		member.Profile.TeamMemberId,
		member.Profile.Email,
		strconv.FormatBool(member.Profile.EmailVerified),
		member.Profile.ExternalId,
		member.Profile.MembershipType.Tag,
		member.Role.Tag,
		member.Profile.Status.Tag,
	}
}
