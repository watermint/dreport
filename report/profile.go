package report

import (
	"encoding/csv"
	"github.com/cihub/seelog"
	"github.com/dropbox/dropbox-sdk-go-unofficial"
	"github.com/dropbox/dropbox-sdk-go-unofficial/team"
	"github.com/watermint/dreport/auth"
	"github.com/watermint/dreport/integration"
	"os"
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

func (t *ReportMemberProfile) Report(context *integration.ExecutionContext) error {
	client := dropbox.Client(context.TeamInfoToken, dropbox.Options{})
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
	members, err := client.MembersList(team.NewMembersListArg())
	if err != nil {
		seelog.Error("Unable to load member list", err)
		return err
	}
	for {
		for _, m := range members.Members {
			if err = t.writeLine(outCsv, m); err != nil {
				seelog.Error("Unable to write data", err)
				return err
			}
		}
		if !members.HasMore {
			seelog.Info("Finished")
			return nil
		}

		seelog.Info("Loading more members..")
		cont := team.NewMembersListContinueArg(members.Cursor)
		members, err = client.MembersListContinue(cont)
		if err != nil {
			seelog.Error("Unable to load member (continue)", err)
			return err
		}
	}
}

func (t *ReportMemberProfile) writeHeader(outCsv *csv.Writer) error {
	header := []string{
		"Account Id",
		"Team Member Id",
		"Email",
		"Email verified?",
		"External Id",
		"Membership Type",
		"Role",
		"Status",
	}

	return outCsv.Write(header)
}

func (t *ReportMemberProfile) writeLine(outCsv *csv.Writer, member *team.TeamMemberInfo) error {
	line := []string{
		member.Profile.AccountId,
		member.Profile.TeamMemberId,
		member.Profile.Email,
		strconv.FormatBool(member.Profile.EmailVerified),
		member.Profile.ExternalId,
		member.Profile.MembershipType.Tag,
		member.Role.Tag,
		member.Profile.Status.Tag,
	}

	return outCsv.Write(line)
}
