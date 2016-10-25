package report

import (
	"encoding/csv"
	"github.com/cihub/seelog"
	"github.com/dropbox/dropbox-sdk-go-unofficial"
	"github.com/dropbox/dropbox-sdk-go-unofficial/team"
	"github.com/dropbox/dropbox-sdk-go-unofficial/users"
	"github.com/watermint/dreport/auth"
	"github.com/watermint/dreport/integration"
	"os"
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

func (t *ReportQuotaUsage) Report(context *integration.ExecutionContext) error {
	infoClient := dropbox.Client(context.TeamInfoToken, dropbox.Options{})
	out, err := os.Create(context.OutputFile)
	if err != nil {
		seelog.Errorf("Unable to create output file: '%s'", context.OutputFile)
		return err
	}
	defer out.Close()
	outCsv := csv.NewWriter(out)
	defer outCsv.Flush()

	if err := t.writeHeader(outCsv); err != nil {
		seelog.Error("Unable to write header line", err)
		return err
	}

	seelog.Info("Loading members")
	members, err := infoClient.MembersList(team.NewMembersListArg())
	if err != nil {
		seelog.Error("Unable to load member list", err)
		return err
	}
	for {
		for _, m := range members.Members {
			memberClient := dropbox.Client(context.TeamFileToken, dropbox.Options{
				AsMemberId: m.Profile.TeamMemberId,
			})

			usage, err := memberClient.GetSpaceUsage()
			if err != nil {
				seelog.Errorf("Unable to load quota for member: '%s'", m.Profile.AccountId)
				return err
			}

			if err = t.writeLine(outCsv, m, usage); err != nil {
				seelog.Errorf("Unable to write data", err)
				return err
			}
		}
		if !members.HasMore {
			seelog.Info("Finished")
			return nil
		}

		seelog.Info("Loading more members..")
		cont := team.NewMembersListContinueArg(members.Cursor)
		members, err = infoClient.MembersListContinue(cont)
		if err != nil {
			seelog.Error("Unable to load member (continue)", err)
			return err
		}
	}

	return nil
}

func (t *ReportQuotaUsage) writeHeader(outCsv *csv.Writer) error {
	header := []string{
		"Account Id",
		"Team Member Id",
		"Email",
		"Usage (bytes)",
	}

	return outCsv.Write(header)
}

func (t *ReportQuotaUsage) writeLine(outCsv *csv.Writer, member *team.TeamMemberInfo, usage *users.SpaceUsage) error {
	line := []string{
		member.Profile.AccountId,
		member.Profile.TeamMemberId,
		member.Profile.Email,
		strconv.FormatUint(usage.Used, 10),
	}

	return outCsv.Write(line)
}
