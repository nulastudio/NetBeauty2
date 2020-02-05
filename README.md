# NetCoreBeauty

## What is it?
Move a .NET Core app runtime components and dependencies into a sub-directory and make it beauty.

## Limitation
Only works with [Self-contained deployments mode](https://docs.microsoft.com/en-us/dotnet/core/deploying/#self-contained-deployments-scd)

目前仅适用于[独立部署发布模式](https://docs.microsoft.com/zh-cn/dotnet/core/deploying/#self-contained-deployments-scd)的程序

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
    <ForceBeauty>False</ForceBeauty>
    <!-- <BeautyAfterTasks></BeautyAfterTasks> -->
    <!-- set to True if you want to disable -->
    <DisablePatch>False</DisablePatch>
    <!-- valid values: Error|Detail|Info -->
    <BeautyLogLevel>Error</BeautyLogLevel>
    <!-- set to a repo mirror if you have troble in connecting github -->
    <!-- <GitCDN>https://gitee.com/liesauer/HostFXRPatcher</GitCDN> -->
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="nulastudio.NetCoreBeauty" />
  </ItemGroup>

</Project>
```
when you run `dotnet publish -r` (only works with `SCD` mode), everything is done automatically.

2. Use the binary application if your project has already be published.
```
Usage:
ncbeauty [--<force=True|False>] [--<gitcdn>] [--<loglevel=Error|Detail|Info>] [--<nopatch=True|False>] <beautyDir> [<libsDir>]
ncbeauty [--<loglevel=Error|Detail|Info>] setcdn <gitcdn>
ncbeauty [--<loglevel=Error|Detail|Info>] getcdn
ncbeauty [--<loglevel=Error|Detail|Info>] delcdn
```
for example
```
ncbeauty /path/to/publishDir
```

3. Install .NETCore Global Tool
```
dotnet tool install --global nulastudio.ncbeauty
```
then use it just like binary distribution.

## Mirror
if you have troble in connecting github, use this mirror
```
https://gitee.com/liesauer/HostFXRPatcher
```

## Default Git CDN
`ncbeauty` [1.2.1](https://github.com/nulastudio/NetCoreBeauty/releases/tag/v1.2.1)  supports setting default Git CDN now, you don't need `--gitcdn` all the time if you are using binary distribution. but how ever default git cdn can be override by `--gitcdn`.
Usage:
```
ncbeauty [--<loglevel=Error|Detail|Info>] setcdn <gitcdn>
  set current default git cdn, can be override by --gitcdn.
ncbeauty [--<loglevel=Error|Detail|Info>] getcdn
  print current default git cdn.
ncbeauty [--<loglevel=Error|Detail|Info>] delcdn
  remove current default git cdn, after removed, use --gitcdn to specify.
```
