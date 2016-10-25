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

func (t *ReportMemberSessions) Report(context *integration.ExecutionContext) error {
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
		seelog.Errorf("Unable to write header line", err)
		return err
	}

	membersMap := make(map[string]*team.TeamMemberInfo)
	seelog.Info("Loading members")
	membersList, err := infoClient.MembersList(team.NewMembersListArg())
	if err != nil {
		seelog.Errorf("Unable to load member list", err)
		return err
	}
	for {
		for _, m := range membersList.Members {
			membersMap[m.Profile.TeamMemberId] = m
		}
		if !membersList.HasMore {
			break
		}
		seelog.Info("Loading more members..")
		cont := team.NewMembersListContinueArg(membersList.Cursor)
		membersList, err = infoClient.MembersListContinue(cont)
		if err != nil {
			seelog.Error("Unable to load member (continue)", err)
			return err
		}
	}

	auditClient := dropbox.Client(context.TeamFileToken, dropbox.Options{})

	seelog.Info("Loading sessions")
	query := team.NewListMembersDevicesArg()
	query.IncludeDesktopClients = true
	query.IncludeMobileClients = true
	query.IncludeWebSessions = true
	sessions, err := auditClient.DevicesListMembersDevices(query)
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
				t.writeDesktopSession(outCsv, member, s)
			}
			for _, s := range d.MobileClients {
				t.writeMobileSession(outCsv, member, s)
			}
			for _, s := range d.WebSessions {
				t.writeWebSession(outCsv, member, s)
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

		sessions, err = auditClient.DevicesListMembersDevices(query)
		if err != nil {
			seelog.Error("Unable to load member (contiue)", err)
		}
	}
}

func (t *ReportMemberSessions) writeHeader(outCsv *csv.Writer) error {
	header := []string{
		"Account Id",
		"Team Member Id",
		"Email",
		"Session Type",
		"Session Id",
		"IP Address",
		"Country",
		"Client Type",
		"Client Version",
		"OS",
		"Platform",
		"OS Version",
		"Last Carrier",
		"Device Name",
		"Hostname",
		"Browser",
		"User agent",
		"Is delete on unlink supported?",
		"Created",
		"Updated",
	}

	return outCsv.Write(header)
}

func (t *ReportMemberSessions) writeDesktopSession(outCsv *csv.Writer, m *team.TeamMemberInfo, s *team.DesktopClientSession) error {
	line := []string{
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

	return outCsv.Write(line)
}

func (t *ReportMemberSessions) writeMobileSession(outCsv *csv.Writer, m *team.TeamMemberInfo, s *team.MobileClientSession) error {
	line := []string{
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

	return outCsv.Write(line)
}

func (t *ReportMemberSessions) writeWebSession(outCsv *csv.Writer, m *team.TeamMemberInfo, s *team.ActiveWebSession) error {
	line := []string{
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

	return outCsv.Write(line)
}
