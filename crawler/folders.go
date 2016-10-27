package crawler

import (
	"github.com/dropbox/dropbox-sdk-go-unofficial"
	"github.com/dropbox/dropbox-sdk-go-unofficial/sharing"
	"github.com/cihub/seelog"
)

func AllSharedFolders(client dropbox.Api) ([]*sharing.SharedFolderMetadata, error) {
	folders := make([]*sharing.SharedFolderMetadata, 0)
	list, err := client.ListFolders(sharing.NewListFoldersArgs())
	if err != nil {
		seelog.Error("Unable to load shared folders", err)
		return nil, err
	}
	for {
		folders = append(folders, list.Entries...)
		if list.Cursor == "" {
			return folders, nil
		}
		list, err = client.ListFoldersContinue(sharing.NewListFoldersContinueArg(list.Cursor))
		if err != nil {
			seelog.Error("Unable to load shared folders (continue)", err)
			return nil, err
		}
	}
}

func AllSharedFolderMembers(client dropbox.Api, sharedFolderId string) ([]*sharing.GroupMembershipInfo, []*sharing.UserMembershipInfo, []*sharing.InviteeMembershipInfo, error) {
	groups := make([]*sharing.GroupMembershipInfo, 0)
	users := make([]*sharing.UserMembershipInfo, 0)
	invitees := make([]*sharing.InviteeMembershipInfo, 0)

	m, err := client.ListFolderMembers(sharing.NewListFolderMembersArgs(sharedFolderId))
	if err != nil {
		seelog.Error("Unable to load shared folder members for shared folder id: " + sharedFolderId, err)
		return nil, nil, nil, err
	}

	for {
		groups = append(groups, m.Groups...)
		users = append(users, m.Users...)
		invitees = append(invitees, m.Invitees...)

		if m.Cursor == "" {
			return groups, users, invitees, nil
		}
		m, err = client.ListFolderMembersContinue(sharing.NewListFolderMembersContinueArg(m.Cursor))
		if err != nil {
			seelog.Error("Unable to load shared folder members (continue) for shared folder id: " + sharedFolderId, err)
			return nil, nil, nil, err
		}
	}
}