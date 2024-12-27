package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gobwas/glob"
	"github.com/samber/lo"
	"golang.org/x/sys/windows"
)

var knownFolderIdMaps = map[string]*windows.KNOWNFOLDERID{
	"NetworkFolder":          windows.FOLDERID_NetworkFolder,
	"ComputerFolder":         windows.FOLDERID_ComputerFolder,
	"InternetFolder":         windows.FOLDERID_InternetFolder,
	"ControlPanelFolder":     windows.FOLDERID_ControlPanelFolder,
	"PrintersFolder":         windows.FOLDERID_PrintersFolder,
	"SyncManagerFolder":      windows.FOLDERID_SyncManagerFolder,
	"SyncSetupFolder":        windows.FOLDERID_SyncSetupFolder,
	"ConflictFolder":         windows.FOLDERID_ConflictFolder,
	"SyncResultsFolder":      windows.FOLDERID_SyncResultsFolder,
	"RecycleBinFolder":       windows.FOLDERID_RecycleBinFolder,
	"ConnectionsFolder":      windows.FOLDERID_ConnectionsFolder,
	"Fonts":                  windows.FOLDERID_Fonts,
	"Desktop":                windows.FOLDERID_Desktop,
	"Startup":                windows.FOLDERID_Startup,
	"Programs":               windows.FOLDERID_Programs,
	"StartMenu":              windows.FOLDERID_StartMenu,
	"Recent":                 windows.FOLDERID_Recent,
	"SendTo":                 windows.FOLDERID_SendTo,
	"Documents":              windows.FOLDERID_Documents,
	"Favorites":              windows.FOLDERID_Favorites,
	"NetHood":                windows.FOLDERID_NetHood,
	"PrintHood":              windows.FOLDERID_PrintHood,
	"Templates":              windows.FOLDERID_Templates,
	"CommonStartup":          windows.FOLDERID_CommonStartup,
	"CommonPrograms":         windows.FOLDERID_CommonPrograms,
	"CommonStartMenu":        windows.FOLDERID_CommonStartMenu,
	"PublicDesktop":          windows.FOLDERID_PublicDesktop,
	"ProgramData":            windows.FOLDERID_ProgramData,
	"CommonTemplates":        windows.FOLDERID_CommonTemplates,
	"PublicDocuments":        windows.FOLDERID_PublicDocuments,
	"RoamingAppData":         windows.FOLDERID_RoamingAppData,
	"LocalAppData":           windows.FOLDERID_LocalAppData,
	"LocalAppDataLow":        windows.FOLDERID_LocalAppDataLow,
	"InternetCache":          windows.FOLDERID_InternetCache,
	"Cookies":                windows.FOLDERID_Cookies,
	"History":                windows.FOLDERID_History,
	"System":                 windows.FOLDERID_System,
	"SystemX86":              windows.FOLDERID_SystemX86,
	"Windows":                windows.FOLDERID_Windows,
	"Profile":                windows.FOLDERID_Profile,
	"Pictures":               windows.FOLDERID_Pictures,
	"ProgramFilesX86":        windows.FOLDERID_ProgramFilesX86,
	"ProgramFilesCommonX86":  windows.FOLDERID_ProgramFilesCommonX86,
	"ProgramFilesX64":        windows.FOLDERID_ProgramFilesX64,
	"ProgramFilesCommonX64":  windows.FOLDERID_ProgramFilesCommonX64,
	"ProgramFiles":           windows.FOLDERID_ProgramFiles,
	"ProgramFilesCommon":     windows.FOLDERID_ProgramFilesCommon,
	"UserProgramFiles":       windows.FOLDERID_UserProgramFiles,
	"UserProgramFilesCommon": windows.FOLDERID_UserProgramFilesCommon,
	"AdminTools":             windows.FOLDERID_AdminTools,
	"CommonAdminTools":       windows.FOLDERID_CommonAdminTools,
	"Music":                  windows.FOLDERID_Music,
	"Videos":                 windows.FOLDERID_Videos,
	"Ringtones":              windows.FOLDERID_Ringtones,
	"PublicPictures":         windows.FOLDERID_PublicPictures,
	"PublicMusic":            windows.FOLDERID_PublicMusic,
	"PublicVideos":           windows.FOLDERID_PublicVideos,
	"PublicRingtones":        windows.FOLDERID_PublicRingtones,
	"ResourceDir":            windows.FOLDERID_ResourceDir,
	"LocalizedResourcesDir":  windows.FOLDERID_LocalizedResourcesDir,
	"CommonOEMLinks":         windows.FOLDERID_CommonOEMLinks,
	"CDBurning":              windows.FOLDERID_CDBurning,
	"UserProfiles":           windows.FOLDERID_UserProfiles,
	"Playlists":              windows.FOLDERID_Playlists,
	"SamplePlaylists":        windows.FOLDERID_SamplePlaylists,
	"SampleMusic":            windows.FOLDERID_SampleMusic,
	"SamplePictures":         windows.FOLDERID_SamplePictures,
	"SampleVideos":           windows.FOLDERID_SampleVideos,
	"PhotoAlbums":            windows.FOLDERID_PhotoAlbums,
	"Public":                 windows.FOLDERID_Public,
	"ChangeRemovePrograms":   windows.FOLDERID_ChangeRemovePrograms,
	"AppUpdates":             windows.FOLDERID_AppUpdates,
	"AddNewPrograms":         windows.FOLDERID_AddNewPrograms,
	"Downloads":              windows.FOLDERID_Downloads,
	"PublicDownloads":        windows.FOLDERID_PublicDownloads,
	"SavedSearches":          windows.FOLDERID_SavedSearches,
	"QuickLaunch":            windows.FOLDERID_QuickLaunch,
	"Contacts":               windows.FOLDERID_Contacts,
	"SidebarParts":           windows.FOLDERID_SidebarParts,
	"SidebarDefaultParts":    windows.FOLDERID_SidebarDefaultParts,
	"PublicGameTasks":        windows.FOLDERID_PublicGameTasks,
	"GameTasks":              windows.FOLDERID_GameTasks,
	"SavedGames":             windows.FOLDERID_SavedGames,
	"Games":                  windows.FOLDERID_Games,
	"SEARCH_MAPI":            windows.FOLDERID_SEARCH_MAPI,
	"SEARCH_CSC":             windows.FOLDERID_SEARCH_CSC,
	"Links":                  windows.FOLDERID_Links,
	"UsersFiles":             windows.FOLDERID_UsersFiles,
	"UsersLibraries":         windows.FOLDERID_UsersLibraries,
	"SearchHome":             windows.FOLDERID_SearchHome,
	"OriginalImages":         windows.FOLDERID_OriginalImages,
	"DocumentsLibrary":       windows.FOLDERID_DocumentsLibrary,
	"MusicLibrary":           windows.FOLDERID_MusicLibrary,
	"PicturesLibrary":        windows.FOLDERID_PicturesLibrary,
	"VideosLibrary":          windows.FOLDERID_VideosLibrary,
	"RecordedTVLibrary":      windows.FOLDERID_RecordedTVLibrary,
	"HomeGroup":              windows.FOLDERID_HomeGroup,
	"HomeGroupCurrentUser":   windows.FOLDERID_HomeGroupCurrentUser,
	"DeviceMetadataStore":    windows.FOLDERID_DeviceMetadataStore,
	"Libraries":              windows.FOLDERID_Libraries,
	"PublicLibraries":        windows.FOLDERID_PublicLibraries,
	"UserPinned":             windows.FOLDERID_UserPinned,
	"ImplicitAppShortcuts":   windows.FOLDERID_ImplicitAppShortcuts,
	"AccountPictures":        windows.FOLDERID_AccountPictures,
	"PublicUserTiles":        windows.FOLDERID_PublicUserTiles,
	"AppsFolder":             windows.FOLDERID_AppsFolder,
	"StartMenuAllPrograms":   windows.FOLDERID_StartMenuAllPrograms,
	"CommonStartMenuPlaces":  windows.FOLDERID_CommonStartMenuPlaces,
	"ApplicationShortcuts":   windows.FOLDERID_ApplicationShortcuts,
	"RoamingTiles":           windows.FOLDERID_RoamingTiles,
	"RoamedTileImages":       windows.FOLDERID_RoamedTileImages,
	"Screenshots":            windows.FOLDERID_Screenshots,
	"CameraRoll":             windows.FOLDERID_CameraRoll,
	"SkyDrive":               windows.FOLDERID_SkyDrive,
	"OneDrive":               windows.FOLDERID_OneDrive,
	"SkyDriveDocuments":      windows.FOLDERID_SkyDriveDocuments,
	"SkyDrivePictures":       windows.FOLDERID_SkyDrivePictures,
	"SkyDriveMusic":          windows.FOLDERID_SkyDriveMusic,
	"SkyDriveCameraRoll":     windows.FOLDERID_SkyDriveCameraRoll,
	"SearchHistory":          windows.FOLDERID_SearchHistory,
	"SearchTemplates":        windows.FOLDERID_SearchTemplates,
	"CameraRollLibrary":      windows.FOLDERID_CameraRollLibrary,
	"SavedPictures":          windows.FOLDERID_SavedPictures,
	"SavedPicturesLibrary":   windows.FOLDERID_SavedPicturesLibrary,
	"RetailDemo":             windows.FOLDERID_RetailDemo,
	"Device":                 windows.FOLDERID_Device,
	"DevelopmentFiles":       windows.FOLDERID_DevelopmentFiles,
	"Objects3D":              windows.FOLDERID_Objects3D,
	"AppCaptures":            windows.FOLDERID_AppCaptures,
	"LocalDocuments":         windows.FOLDERID_LocalDocuments,
	"LocalPictures":          windows.FOLDERID_LocalPictures,
	"LocalVideos":            windows.FOLDERID_LocalVideos,
	"LocalMusic":             windows.FOLDERID_LocalMusic,
	"LocalDownloads":         windows.FOLDERID_LocalDownloads,
	"RecordedCalls":          windows.FOLDERID_RecordedCalls,
	"AllAppMods":             windows.FOLDERID_AllAppMods,
	"CurrentAppMods":         windows.FOLDERID_CurrentAppMods,
	"AppDataDesktop":         windows.FOLDERID_AppDataDesktop,
	"AppDataDocuments":       windows.FOLDERID_AppDataDocuments,
	"AppDataFavorites":       windows.FOLDERID_AppDataFavorites,
	"AppDataProgramData":     windows.FOLDERID_AppDataProgramData,
}

var patterns = []string{
	// Windows
	"$RECYCLE.BIN/",
	"Config.Msi/",
	"FOUND.[0-9][0-9][0-9]/",
	"System Volume Information/",
	"DumpStack.log*",
	"**/*.ink",
	"**/*.stackdump",
	"**/Desktop.ini",
	"**/Thumbs.db",
	"**/Thumbs.db:encryptable",
	"**/ehthumbs.db",
	"**/ehthumbs_vista.db",
	// macOS
	".DocumentRevisions-V100/",
	".Spotlight-V100/",
	".TemporaryItems/",
	".Trashes/",
	".VolumeIcon.icns/",
	".com.apple.timemachine.donotpresent/",
	".fseventsd/",
	"**/.AppleDB/",
	"**/.AppleDesktop/",
	"**/.AppleDouble/",
	"**/.DS_Store/",
	"**/.LSOverride/",
	"**/.apdisk/",
	"**/__MACOSX/",
	"**/*.icloud",
	"**/._*",
	// Linux
	"**/.Trash-*/",
	"**/*~",
	"**/.fuse_hidden*",
	"**/.nfs*",
	// Others
	"@Recently-Snapshot/",
	"@Recycle/",
	"**/*.tmp",
	"**/~$*",
}

func getPartitionLabels() ([]rune, error) {
	drivesBitMask, err := windows.GetLogicalDrives()
	if err != nil {
		return nil, err
	}
	labels := make([]rune, 0, 26)
	for r := 'A'; r <= 'Z'; r++ {
		if drivesBitMask&1 != 0 {
			labels = append(labels, r)
		}
		drivesBitMask >>= 1
	}
	return labels, nil
}

func getUserDirs() ([]string, error) {
	dirs := make([]string, 0, len(knownFolderIdMaps))
	for _, id := range knownFolderIdMaps {
		dir, err := windows.KnownFolderPath(id, windows.KF_FLAG_DEFAULT)
		if err != nil {
			if errors.Is(err, syscall.Errno(windows.E_FAIL)) {
				continue
			}
			if errors.Is(err, syscall.Errno(0x80070000|windows.ERROR_FILE_NOT_FOUND)) {
				continue
			}
			if errors.Is(err, syscall.Errno(0x80070000|windows.ERROR_PATH_NOT_FOUND)) {
				continue
			}
			return nil, err
		}
		dirs = append(dirs, dir)
	}
	return dirs, nil
}

func getAllDirsAndFiles(rootDir string) ([]string, []string, error) {
	dirs := make([]string, 0)
	files := make([]string, 0)
	err := filepath.WalkDir(
		filepath.Clean(rootDir),
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if errors.Is(err, syscall.Errno(windows.ERROR_NOT_READY)) {
					return nil
				}
				if errors.Is(err, fs.ErrPermission) {
					return nil
				}
				return err
			}
			if d.IsDir() {
				dirs = append(dirs, filepath.ToSlash(path)+"/")
			} else {
				files = append(files, filepath.ToSlash(path))
			}
			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}
	return dirs, files, nil
}

func main() {
	rootDirs := lo.FilterMap(
		must1(getPartitionLabels()),
		func(label rune, index int) (string, bool) {
			if label == 'C' {
				return "", false
			}
			return string(label) + ":/", true
		},
	)
	userDirs := lo.FilterMap(
		must1(getUserDirs()),
		func(dir string, index int) (string, bool) {
			isRel := lo.ContainsBy(rootDirs, func(rootDir string) bool {
				relPath, _ := filepath.Rel(rootDir, dir)
				relPath = filepath.ToSlash(relPath)
				return relPath != ""
			})
			return filepath.ToSlash(dir) + "/", isRel
		},
	)
	if len(os.Args) > 1 {
		p := os.Args[1]
		if p[len(p)-1:] != "/" || p[len(p)-1:] != "\\" {
			p += "/"
		}
		p = filepath.ToSlash(filepath.Clean(p))
		if p[len(p)-1:] != "/" || p[len(p)-1:] != "\\" {
			p += "/"
		}
		rootDirs = []string{p}
	}

	matchedDirs := make([]string, 0)
	matchedFiles := make([]string, 0)
	for _, rootDir := range rootDirs {
		matcherMap := lo.SliceToMap(patterns, func(pattern string) (string, glob.Glob) {
			return pattern, glob.MustCompile(strings.ToLower(rootDir+pattern), '/')
		})
		dirs, files, err := getAllDirsAndFiles(rootDir)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for pattern, matcher := range matcherMap {
			if pattern[len(pattern)-1:] == "/" {
				for _, dir := range dirs {
					if matcher.Match(strings.ToLower(dir)) {
						matchedDirs = append(matchedDirs, dir)
					}
				}
			} else {
				for _, file := range files {
					if matcher.Match(strings.ToLower(file)) {
						if pattern == "**/Desktop.ini" {
							inUserDir := lo.ContainsBy(userDirs, func(userDir string) bool {
								relPath, _ := filepath.Rel(userDir, filepath.Dir(file))
								relPath = filepath.ToSlash(relPath)
								return relPath == "."
							})
							if inUserDir {
								continue
							}
						}
						matchedFiles = append(matchedFiles, file)
					}
				}
			}
		}
	}

	fmt.Println("Matched files:")
	for _, file := range matchedFiles {
		fmt.Println("-", file)
	}
	fmt.Println("Matched directories:")
	for _, dir := range matchedDirs {
		fmt.Println("-", dir)
	}

	fmt.Print("Do you want to delete these files and directories? (y/n): ")
	var input string
	fmt.Scanln(&input)
	if strings.ToLower(input) != "y" {
		return
	}

	for _, file := range matchedFiles {
		fmt.Printf("Deleting '%s' ...\n", file)
		err := os.Remove(file)
		if err != nil {
			fmt.Println(err)
		}
	}
	for _, dir := range matchedDirs {
		fmt.Printf("Deleting '%s' ...\n", dir)
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func must1[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
