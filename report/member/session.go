package member

import (
	"github.com/cihub/seelog"
	"github.com/dropbox/dropbox-sdk-go-unofficial"
	"github.com/dropbox/dropbox-sdk-go-unofficial/team"
	"github.com/watermint/dreport/auth"
	"github.com/watermint/dreport/crawler"
	"github.com/watermint/dreport/integration"
	"strconv"
)

type ReportMemberSessions struct {
}

func (t *ReportMemberSessions) ReportName() string {
	return "TeamMemberSession"
}

func (t *ReportMemberSessions) ReportDescription() string {
	return "List existing sessions (desktop/mobile/web) of all team member of a team"
}

func (t *ReportMemberSessions) RequiredPermissions() []string {
	return []string{
		auth.PERMISSION_INFO,
		auth.PERMISSION_FILE,
	}
}

func (t *ReportMemberSessions) Report(context *integration.ReportContext) error {
	members, err := crawler.AllTeamMembers(context)
	if err != nil {
		seelog.Errorf("Unable to load member list", err)
		return err
	}
	membersMap := make(map[string]*team.TeamMemberInfo)
	for _, m := range members {
		membersMap[m.Profile.TeamMemberId] = m
	}

	fileClient := dropbox.Client(context.TeamFileToken, dropbox.Options{})

	context.ReportOutput.Headers(t.createHeader())

	seelog.Info("Loading sessions")
	query := team.NewListMembersDevicesArg()
	query.IncludeDesktopClients = true
	query.IncludeMobileClients = true
	query.IncludeWebSessions = true
	sessions, err := fileClient.DevicesListMembersDevices(query)
	if err != nil {
		seelog.Error("Unable to load members sessions", err)
		return err
	}
	for {
		for _, d := range sessions.Devices {
			member, found := membersMap[d.TeamMemberId]
			if !found {
				seelog.Errorf("Member profile not found for Team Member Id: %s", d.TeamMemberId)
				continue
			}
			for _, s := range d.DesktopClients {
				context.ReportOutput.Row(t.createDesktopSession(member, s))
			}
			for _, s := range d.MobileClients {
				context.ReportOutput.Row(t.createMobileSession(member, s))
			}
			for _, s := range d.WebSessions {
				context.ReportOutput.Row(t.createWebSession(member, s))
			}
		}
		if !sessions.HasMore {
			seelog.Info("Finished")
			return nil
		}

		seelog.Info("Loading more sessions..")
		query = team.NewListMembersDevicesArg()
		query.Cursor = sessions.Cursor
		query.IncludeDesktopClients = true
		query.IncludeMobileClients = true
		query.IncludeWebSessions = true

		sessions, err = fileClient.DevicesListMembersDevices(query)
		if err != nil {
			seelog.Error("Unable to load member (contiue)", err)
		}
	}
}

func (t *ReportMemberSessions) createHeader() []string {
	return []string{
		"account-id",
		"team-member-id",
		"email",
		"session-type",
		"session-id",
		"ip-address",
		"country",
		"client-type",
		"client-version",
		"os",
		"platform",
		"os-version",
		"last-carrier",
		"device-name",
		"hostname",
		"browser",
		"user-agent",
		"is-delete-on-unlink-supported",
		"created",
		"updated",
	}
}

func (t *ReportMemberSessions) createDesktopSession(m *team.TeamMemberInfo, s *team.DesktopClientSession) []string {
	return []string{
		m.Profile.AccountId,
		m.Profile.TeamMemberId,
		m.Profile.Email,
		"Desktop",
		s.SessionId,
		s.IpAddress,
		s.Country,
		s.ClientType.Tag,
		s.ClientVersion,
		"", // OS
		s.Platform,
		"", // OS version
		"", // Last carrier
		"", // Device name
		s.HostName,
		"", // Browser
		"", // User agent
		strconv.FormatBool(s.IsDeleteOnUnlinkSupported),
		s.Created.String(),
		s.Updated.String(),
	}
}

func (t *ReportMemberSessions) createMobileSession(m *team.TeamMemberInfo, s *team.MobileClientSession) []string {
	return []string{
		m.Profile.AccountId,
		m.Profile.TeamMemberId,
		m.Profile.Email,
		"Mobile",
		s.SessionId,
		s.IpAddress,
		s.Country,
		s.ClientType.Tag,
		s.ClientVersion,
		"", // OS
		"", // Platform
		s.OsVersion,
		s.LastCarrier,
		s.DeviceName,
		"", // Hostname
		"", // Browser
		"", // User agent
		"", // Is delete on unlink supported
		s.Created.String(),
		s.Updated.String(),
	}
}

func (t *ReportMemberSessions) createWebSession(m *team.TeamMemberInfo, s *team.ActiveWebSession) []string {
	return []string{
		m.Profile.AccountId,
		m.Profile.TeamMemberId,
		m.Profile.Email,
		"Web",
		s.SessionId,
		s.IpAddress,
		s.Country,
		"", // Client type
		"", // Client version
		s.Os,
		"", // Platform
		"", // OS version
		"", // Last carrier
		"", // device name
		"", // hostname
		s.Browser,
		s.UserAgent,
		"", // Is delete on unlink supported
		s.Created.String(),
		s.Updated.String(),
	}
}
