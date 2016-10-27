package crawler

import (
	"github.com/cihub/seelog"
	"github.com/dropbox/dropbox-sdk-go-unofficial"
	"github.com/dropbox/dropbox-sdk-go-unofficial/team"
	"github.com/watermint/dreport/integration"
)

func AllTeamMembers(ctx *integration.ReportContext) ([]*team.TeamMemberInfo, error) {
	memberList := make([]*team.TeamMemberInfo, 0, 0)
	client := dropbox.Client(ctx.TeamInfoToken, dropbox.Options{})

	seelog.Info("Loading members")
	members, err := client.MembersList(team.NewMembersListArg())
	if err != nil {
		seelog.Error("Unable to load member list", err)
		return memberList, err
	}
	for {
		memberList = append(memberList, members.Members...)
		if !members.HasMore {
			seelog.Info("Finished loading member list")
			return memberList, nil
		}
		seelog.Info("Loading more members..")
		cont := team.NewMembersListContinueArg(members.Cursor)
		members, err = client.MembersListContinue(cont)
		if err != nil {
			seelog.Error("Unable to load member (continue)", err)
			return memberList, err
		}
	}

}
