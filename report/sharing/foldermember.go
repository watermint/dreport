package sharing

import (
	"github.com/watermint/dreport/auth"
	"github.com/watermint/dreport/integration"
	"github.com/watermint/dreport/crawler"
	"github.com/dropbox/dropbox-sdk-go-unofficial/sharing"
	"github.com/dropbox/dropbox-sdk-go-unofficial"
	"github.com/cihub/seelog"
	"strconv"
)

type ReportSharedFolderMembers struct {
}

func (t *ReportSharedFolderMembers) ReportName() string {
	return "SharedFolderMembers"
}

func (t *ReportSharedFolderMembers) ReportDescription() string {
	return "List all shared folders and their members of a team"
}

func (t *ReportSharedFolderMembers) RequiredPermissions() []string {
	return []string{
		auth.PERMISSION_INFO,
		auth.PERMISSION_FILE,
	}
}

func (t *ReportSharedFolderMembers) Report(rc *integration.ReportContext) error {
	members, err := crawler.AllTeamMembers(rc)
	if err != nil {
		return err
	}

	rc.ReportOutput.Headers(t.createHeader())

	// Load all shared folders
	sharedFolders := make(map[string]*sharing.SharedFolderMetadata)
	sharedFolderAsMember := make(map[string]string)
	for _, m := range members {
		client := dropbox.Client(rc.TeamFileToken, dropbox.Options{
			AsMemberId: m.Profile.TeamMemberId,
		})
		folders, err := crawler.AllSharedFolders(client)
		if err != nil {
			seelog.Errorf("Unable to load shared folders for member (%s)", m.Profile.TeamMemberId)
			return err
		}
		for _, f := range folders {
			sharedFolders[f.SharedFolderId] = f
			sharedFolderAsMember[f.SharedFolderId] = m.Profile.TeamMemberId
		}
	}

	// Load shared folder members
	for sfid, sf := range sharedFolders {
		asMember, found := sharedFolderAsMember[sfid]
		if !found {
			seelog.Warnf("Unexpected condition. Could not determine 'AsMemberId' for shared folder '%s'", sfid)
			continue
		}
		client := dropbox.Client(rc.TeamFileToken, dropbox.Options{
			AsMemberId: asMember,
		})
		groups, users, invitees, err := crawler.AllSharedFolderMembers(client, sfid)
		if err != nil {
			seelog.Warnf("Unable to load shared folder member information for shared folder '%s'", sfid)
			continue
		}

		for _, g := range groups {
			rc.ReportOutput.Row(t.createGroupRow(sf, g))
		}
		for _, u := range users {
			rc.ReportOutput.Row(t.createUserRow(sf, u))
		}
		for _, i := range invitees {
			rc.ReportOutput.Row(t.createInviteeRow(sf, i))
		}
	}

	return nil
}

func (t *ReportSharedFolderMembers) createHeader() []string {
	return []string{
		"shared-folder-id",
		"shared-folder-name",
		"is-team-folder",
		"management-type",
		"access-level",
		"account-id",
		"team-member-id",
		"email",
		"same-team",
		"group-id",
		"group-external-id",
		"group-name",
	}
}

func (t *ReportSharedFolderMembers) createGroupRow(sf *sharing.SharedFolderMetadata, g *sharing.GroupMembershipInfo) []string {
	return []string{
		sf.SharedFolderId,
		sf.Name,
		strconv.FormatBool(sf.IsTeamFolder),
		"group",
		g.AccessType.Tag,
		"", // account-id
		"", // team-member-id
		"", // email
		"", // same-team
		g.Group.GroupId,
		g.Group.GroupExternalId,
		g.Group.GroupName,
	}
}

func (t *ReportSharedFolderMembers) createUserRow(sf *sharing.SharedFolderMetadata, u *sharing.UserMembershipInfo) []string {
	return []string{
		sf.SharedFolderId,
		sf.Name,
		strconv.FormatBool(sf.IsTeamFolder),
		"user",
		u.AccessType.Tag,
		u.User.AccountId,
		u.User.TeamMemberId,
		"", // email
		strconv.FormatBool(u.User.SameTeam),
		"", // group-id
		"", // group-external-id
		"", // group-name
	}
}

func (t *ReportSharedFolderMembers) createInviteeRow(sf *sharing.SharedFolderMetadata, i *sharing.InviteeMembershipInfo) []string {
	userAccountId := ""
	userTeamMemberId := ""
	userSameTeam := ""

	if i.User != nil {
		userAccountId = i.User.AccountId
		userTeamMemberId = i.User.TeamMemberId
		userSameTeam = strconv.FormatBool(i.User.SameTeam)
	}

	return []string{
		sf.SharedFolderId,
		sf.Name,
		strconv.FormatBool(sf.IsTeamFolder),
		"invitee",
		i.AccessType.Tag,
		userAccountId,
		userTeamMemberId,
		i.Invitee.Email,
		userSameTeam,
		"", // group-id
		"", // group-external-id
		"", // group-name
	}
}