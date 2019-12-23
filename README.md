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
ncbeauty [--<gitcdn>] [--<loglevel=Error|Detail|Info>] [--<nopatch=True|False>] <beautyDir> [<libsDir>]
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
