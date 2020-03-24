# NetCoreBeauty

## What is it?
Move a .NET Core app runtime components and dependencies into a sub-directory and make it beauty.

### After Beauty
![after_beauty](after_beauty.png)

### Before Beauty
![before_beauty](before_beauty.png)

## Why And Why Not?
1. WHY NOT [Single-file Publish](https://docs.microsoft.com/en-us/dotnet/core/whats-new/dotnet-core-3-0#single-file-executables)?

   First, you cannot use it in `.Net Core 2.x` obviously. Second, single-file app will extract everything into temporary directory including apphost, it means that portable app is impossible if you are using `Assembly.GetEntryAssembly().Location` to storage datas.

2. WHY NOT [Fody/Costura](https://github.com/Fody/Costura)?

   [Fody/Costura](https://github.com/Fody/Costura) is fantastic project. But there are some reasons that why i don't want to use it. First, it modifies IL code which will have chances to break you app. Second, you need to change lots of things in order to make assemblies become `embedded resources`, and in some cases, this will break you app. Actually, it broke one of my apps with no reason once. Third, it changes something and complicated you entire project and you cannot know what exactly be changed, it makes weird BUGs and makes you mad. It made me mad once and i am not going to use it anymore. I don't like the feeling that there's no clue at all.

3. WHY NOT [Warp](https://github.com/dgiagio/warp)?

   Like `Single-file Publish`, it does not extract datas into `APP_BASE`, it extracts into `%APPDATA%`, portable app is impossible either. 
In addition, `Warp` cannot set the icon or assembly informations and it won't extract and reuse those infos from the original app.

4. WHY NOT [ILMerge](https://github.com/dotnet/ILMerge)?

   It merges multiple assemblies into a single assembly. So it need to modify your assemblies, and it is not easy to use either. But you still can re-sign all your assemblies so it won't lose strong name.

5. WHY NOT [AppHostPatcher](https://github.com/0xd4d/dnSpy/tree/master/Build/AppHostPatcher)?

   Same goal with `ncbeauty`, but has a little problem. Datas are storaged inside `APP_BASE/sub-dir`, not storage into `APP_BASE` directly, that is because `apphost`'s main assembly has been moved into `APP_BASE/sub-dir`, the actual `APP_BASE` has changed to `APP_BASE/sub-dir`.

6. WHY [NetCoreBeauty](https://github.com/nulastudio/NetCoreBeauty)?

   Simple and single goal, simple and single function. It does nothing to your project and assemblies, it only organizes your app's directory, so you need to do nothing but reference the NuGet package.

## How?
Theoretically, loading assemblies from a subdirectory should be native supported([see `additionalProbingPaths` setting under `runtimeOptions`](https://github.com/dotnet/toolset/blob/master/Documentation/specs/runtime-configuration-file.md#runtimeoptions-section-runtimeconfigjson)), but setting `additionalProbingPaths` in `.runtimeconfig.json` has a serious problem, the host does not resolve relative path from `APP_BASE` but current working directory, therefore we cannot execute the app outside the `APP_BASE`, it means that the only way to run the app is via command `cd APP_BASE & ./executable`, double-click to run is impossible. So i create [HostFXRPatcher](https://github.com/nulastudio/HostFXRPatcher) and fix this problem(just let you gays know, there are several related but difference issues out there, `NetCoreBeauty` does lots of tricks to make it happen), then rebuild the corehost. When publish, `ncbeauty` will try to download the specific patched hostfxr(that is why `ncbeauty` only works with [self-contained deployments mode](https://docs.microsoft.com/en-us/dotnet/core/deploying/#self-contained-deployments-scd)) and modify `.runtimeconfig.json` and `.deps.json`. It is tough to achieve, but `ncbeauty` has made it, that's the point. Why not PR? Because this fix breaks lots of things, merge is not going to happen in a short time. `.NET` community already plan to fix in `.NET 5`.

## Limitation
Only works with [Self-contained deployments mode](https://docs.microsoft.com/en-us/dotnet/core/deploying/#self-contained-deployments-scd)

目前仅适用于[独立部署发布模式](https://docs.microsoft.com/zh-cn/dotnet/core/deploying/#self-contained-deployments-scd)的程序

## Supported OS
OS      | Architectures
--------|--------------
Windows | x64, x86
Linux   | x64
MacOS   | x64

## Change Log
see [Change LOG.md](CHANGELOG.md)

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
    <!-- <GitTree>master</GitTree> -->
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
ncbeauty [--<force=True|False>] [--<gitcdn>] [--<gittree>] [--<loglevel=Error|Detail|Info>] [--<nopatch=True|False>] <beautyDir> [<libsDir>]
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
`ncbeauty` [1.2.1](https://github.com/nulastudio/NetCoreBeauty/releases/tag/v1.2.1) supports setting default Git CDN now, you don't need `--gitcdn` all the time if you are using binary distribution. but how ever default git cdn can be override by `--gitcdn`.
Usage:
```
ncbeauty [--<loglevel=Error|Detail|Info>] setcdn <gitcdn>
  set current default git cdn, can be override by --gitcdn.
ncbeauty [--<loglevel=Error|Detail|Info>] getcdn
  print current default git cdn.
ncbeauty [--<loglevel=Error|Detail|Info>] delcdn
  remove current default git cdn, after removed, use --gitcdn to specify.
```

## Git Tree
Use `--gittree` to specify a valid git branch or any bits commit hash(up to 40) to grab the specific artifacts and won't get updates any more.
default is master, means that you always use the latest artifacts.

NOTE: please provide as longer commit hash as you can, otherwise it may can not be determined as a valid unique commit hash.

NOTE: PLEASE DO NOT USE ANY COMMIT THAT OLDER THEN `995a9774a75975510b352c1935e232c9e2d5b190`

examples:
```
master
feature/xxx
995a977
995a9774a7
995a9774a75975510b352c1935e232c9e2d5b190
```
