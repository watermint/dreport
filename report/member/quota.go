package member

import (
	"github.com/cihub/seelog"
	"github.com/dropbox/dropbox-sdk-go-unofficial"
	"github.com/dropbox/dropbox-sdk-go-unofficial/team"
	"github.com/dropbox/dropbox-sdk-go-unofficial/users"
	"github.com/watermint/dreport/auth"
	"github.com/watermint/dreport/crawler"
	"github.com/watermint/dreport/integration"
	"strconv"
)

type ReportQuotaUsage struct {
}

func (t *ReportQuotaUsage) ReportName() string {
	return "TeamMemberQuota"
}

func (t *ReportQuotaUsage) ReportDescription() string {
	return "List storage usage of all team member of a team"
}

func (t *ReportQuotaUsage) RequiredPermissions() []string {
	return []string{
		auth.PERMISSION_INFO,
		auth.PERMISSION_FILE,
	}
}

func (t *ReportQuotaUsage) Report(context *integration.ReportContext) error {
	members, err := crawler.AllTeamMembers(context)
	if err != nil {
		return err
	}
	context.ReportOutput.Headers(t.createHeader())

	for _, m := range members {
		memberClient := dropbox.Client(context.TeamFileToken, dropbox.Options{
			AsMemberId: m.Profile.TeamMemberId,
		})

		usage, err := memberClient.GetSpaceUsage()
		if err != nil {
			seelog.Errorf("Unable to load quota for member: '%s'", m.Profile.AccountId)
			return err
		}

		context.ReportOutput.Row(t.createRow(m, usage))
	}

	return nil
}

func (t *ReportQuotaUsage) createHeader() []string {
	return []string{
		"Account Id",
		"Team Member Id",
		"Email",
		"Usage (bytes)",
	}
}

func (t *ReportQuotaUsage) createRow(member *team.TeamMemberInfo, usage *users.SpaceUsage) []string {
	return []string{
		member.Profile.AccountId,
		member.Profile.TeamMemberId,
		member.Profile.Email,
		strconv.FormatUint(usage.Used, 10),
	}
}
