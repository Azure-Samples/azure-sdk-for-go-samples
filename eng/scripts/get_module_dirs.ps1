Param(
  [string] $rootDir
)

$modDirs = @()

$pkgFiles = Get-ChildItem -Path $rootDir -Include "go.mod" -Recurse
foreach ($pkgFile in $pkgFiles)
{
  $modDirs += $pkgFile.DirectoryName
}

return $modDirs
