; Installer wizard "Next-Next-Finish" - lihat ROADMAP-windows-desktop.md untuk
; alasan pemilihan Inno Setup (bukan WiX/.msi) dan keputusan tanpa code-signing.
; Dibuild oleh CI (.github/workflows/build-windows.yml) lewat ISCC.exe, bukan manual.

#define AppName "Parse Bankers"
#define AppExeName "pilot-diff.exe"

[Setup]
AppName={#AppName}
AppVersion=1.0.0
DefaultDirName={autopf}\{#AppName}
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
OutputDir=dist
OutputBaseFilename=ParseBankers-Setup
Compression=lzma
SolidCompression=yes
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64
SetupIconFile=..\backend\winres\icon.ico
UninstallDisplayIcon={app}\{#AppExeName}

[Files]
Source: "..\backend\pilot-diff.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\backend\WebView2Loader.dll"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#AppName}"; Filename: "{app}\{#AppExeName}"
Name: "{autodesktop}\{#AppName}"; Filename: "{app}\{#AppExeName}"

[Run]
Filename: "{app}\{#AppExeName}"; Description: "Jalankan {#AppName}"; Flags: nowait postinstall skipifsilent
