# NetCoreBeauty

## What is it?
Move a .NET Core app runtime components and dependencies into a sub-directory and make it beauty.

## Before Beauty
![before_beauty](before_beauty.png)

## After Beauty
![after_beauty](after_beauty.png)

## How to use?
1. Add Nuget reference into your .NET Core project.
```
dotnet add package nulastudio.NetCoreBeauty
```
your `*.csproj` should be similar like this
```xml
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>netcoreapp2.1</TargetFramework>
    <!-- beauty into sub-directory, default is libs, quote with "" if contains space  -->
    <BeautyLibsDir>runtimes</BeautyLibsDir>
    <!-- set to True if you want to disable -->
    <DisableBeauty>False</DisableBeauty>
    <!-- <BeautyAfterTasks></BeautyAfterTasks> -->
    <!-- set to True if you want to disable -->
    <DisablePatch>False</DisablePatch>
    <!-- valid values: Error|Detail|Log -->
    <BeautyLogLevel>Error</BeautyLogLevel>
    <!-- set to a repo mirror if you have troble in connecting github -->
    <!-- <GitCDN>https://github-like.com/someone/HostFXRPatcherMirror</GitCDN> -->
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="nulastudio.NetCoreBeauty" Version="1.1.1" />
  </ItemGroup>

</Project>
```
when you run `dotnet publish` , everything is done automatically.

2. Use the binary application if your project has already be published.
```
Usage:
ncbeauty [--<gitcdn>] [--<loglevel=Error|Detail|Log>] [--<nopatch=True|False>] <beautyDir> [<libsDir>]
```
for example
```
ncbeauty /path/to/publishDir
```

## Mirror
if you have troble in connecting github, use this mirror
```
https://gitee.com/liesauer/HostFXRPatcher
```
