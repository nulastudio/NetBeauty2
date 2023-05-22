# NetBeauty 2

## What is it?
NetBeauty moves a .NET Framework/.NET Core app runtime components and dependencies into a sub-directory and make it beauty.

### After Beauty
![after_beauty](screenshot/after_beauty.png)

**EVEN LESS!**

see [`--hiddens`](#use-the-binary-application-if-your-project-has-already-been-published) option

![after_beauty_with_hiddens](screenshot/after_beauty_with_hiddens.png)

### Before Beauty
![before_beauty](screenshot/before_beauty.png)

## What's New?
|  | [NetBeauty 2](https://github.com/nulastudio/NetBeauty2) | [NetCoreBeauty](https://github.com/nulastudio/NetBeauty2/tree/v1) |
| ---- | ---- | ---- |
| Supported Framework | `.Net Framework`<br/>`.Net Core 3.0+` | `.Net Core 2.0+` |
| Supported Deployment Model | Framework-dependent deployment (`FDD`)<br/>Self-contained deployment (`SCD`)<br/>Framework-dependent executables (`FDE`) | Self-contained deployment (`SCD`) |
| Supported System | All | `win-x64` `win-x86`<br/>`linux-x64` `linux-arm` `linux-arm64`<br/>`osx-x64` |
| Need Patched HostFXR | No<br />Yes(if use patch) | Yes |
| Minimum Structure | ~20 Files<br />~8 Files(if use patch) | ~8 Files |
| How It Works | [`STARTUP_HOOKS`](https://github.com/dotnet/runtime/blob/main/docs/design/features/host-startup-hook.md)<br/>[`AssemblyLoadContext.Resolving`](https://docs.microsoft.com/en-us/dotnet/api/system.runtime.loader.assemblyloadcontext.resolving?view=netcore-3.0)<br/>[`AssemblyLoadContext.ResolvingUnmanagedDll`](https://docs.microsoft.com/en-us/dotnet/api/system.runtime.loader.assemblyloadcontext.resolvingunmanageddll?view=netcore-3.0)<br />+<br />[`patched libhostfxr`](https://github.com/nulastudio/HostFXRPatcher)(if use patch)<br/>[`additionalProbingPaths`](https://github.com/dotnet/toolset/blob/master/Documentation/specs/runtime-configuration-file.md#runtimeoptions-section-runtimeconfigjson)(if use patch) | [`patched libhostfxr`](https://github.com/nulastudio/HostFXRPatcher)<br/>[`additionalProbingPaths`](https://github.com/dotnet/toolset/blob/master/Documentation/specs/runtime-configuration-file.md#runtimeoptions-section-runtimeconfigjson) |
| Shared Runtime | Yes | Possible If Using `patched libhostfxr` Alone |

## The patch is back!
One of the main goals of NetBeauty2 is trying to use a customize loader to replace the patch, but in fact, the loader need to use lots of Types and APIs like `Dictionary<TKey, TValue>` `List<T>` `Path.GetFullPath` `File.Exists` `NativeLibrary` `RuntimeInformation` etc. this causes lots of assembly references and the worst thing is that those files can not be moved, otherwise CoreCLR will failed to initialize and invoke the loader. More complex logic, more files. So i have to make it back.

Now they work excellently together!

the loader lets us support `FDD`/`FDE` apps.<br />
the patch reduces the file count as possible(`SCD` app only).

## How to use?
### Add Nuget reference to your .NET Core project.
```
dotnet add package nulastudio.NetBeauty
```
Your `*.csproj` should be like:
```xml
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>netcoreapp3.0</TargetFramework>
  </PropertyGroup>

  <PropertyGroup>
    <BeautySharedRuntimeMode>False</BeautySharedRuntimeMode>
    <!-- beauty into sub-directory, default is libs, quote with "" if contains space  -->
    <BeautyLibsDir Condition="$(BeautySharedRuntimeMode) == 'True'">../libraries</BeautyLibsDir>
    <BeautyLibsDir Condition="$(BeautySharedRuntimeMode) != 'True'">./libraries</BeautyLibsDir>
    <!-- dlls that you don't want to be moved or can not be moved -->
    <!-- <BeautyExcludes>dll1.dll;lib*;...</BeautyExcludes> -->
    <!-- dlls that end users never needed, so hide them -->
    <!-- <BeautyHiddens>hostfxr;hostpolicy;*.deps.json;*.runtimeconfig*.json</BeautyHiddens> -->
    <!-- set to True if you want to disable -->
    <DisableBeauty>False</DisableBeauty>
    <!-- set to False if you want to beauty on build -->
    <BeautyOnPublishOnly>False</BeautyOnPublishOnly>
    <!-- set to True if you want to allow 3rd debuggers(like dnSpy) debugs the app -->
    <BeautyEnableDebugging>False</BeautyEnableDebugging>
    <!-- the patch can reduce the file count -->
    <!-- set to False if you want to disable -->
    <!-- SCD Mode Feature Only -->
    <BeautyUsePatch>True</BeautyUsePatch>
    <!-- <BeautyAfterTasks></BeautyAfterTasks> -->
    <!-- valid values: Error|Detail|Info -->
    <BeautyLogLevel>Info</BeautyLogLevel>
    <!-- set to a repo mirror if you have troble in connecting github -->
    <!-- <BeautyGitCDN>https://gitee.com/liesauer/HostFXRPatcher</BeautyGitCDN> -->
    <!-- <BeautyGitTree>master</BeautyGitTree> -->
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="nulastudio.NetBeauty" Version="2.1.3.3" />
  </ItemGroup>

</Project>
```
When you run `dotnet build` or `dotnet publish`, everything will be done automatically.

### Use the binary application if your project has already been published.
```
Usage:
nbeauty2 [--srmode] [--usepatch] [--enabledebug] [--loglevel=(Error|Detail|Info)] [--hiddens=<HiddenFiles>] [--roll-forward=<rollForward>] [--gitcdn=<GitCDN>] [--gittree=<GitTree>] <beautyDir> [<libsDir> [<excludes>]]
```

for example
```
ncbeauty2 --usepatch --loglevel Detail --hiddens "hostfxr;hostpolicy;*.deps.json;*.runtimeconfig*.json" /path/to/publishDir libraries "dll1.dll;lib*;..."
```


**`--hiddens` option just hiding the files, not move them, and only works under Windows!**


### Install as a .NETCore Global Tool
```
dotnet tool install --global nulastudio.nbeauty
```
then use it just like normal binary distribution.

## Shared Runtime Structure
```
├── libraries                   - shared runtime dlls(customizable name)
│   ├── locales                 - satellite assemblies
│   │   ├── en
│   │   │   └── *.resources.dll
│   │   │       ├── MD5_1       - allows multiple runtimes between apps.
│   │   │       │   └── *.resources.dll
│   │   │       └── MD5_2
│   │   │           └── *.resources.dll
│   │   │
│   │   ├── zh-Hans
│   │   │   └── *.resources.dll
│   │   │       ├── MD5_1
│   │   │       │   └── *.resources.dll
│   │   │       └── MD5_2
│   │   │           └── *.resources.dll
│   │   │
│   │   └── ...                 - others languages
│   │
│   ├── *.dll                   - shared managed assemblies
│   │   ├── MD5_1
│   │   │   └── *.dll
│   │   └── MD5_2
│   │       └── *.dll
│   │
│   └── srm_native              - native dlls(can't be shared, each app has a full
│       ├── APPID_1               copy of their own native dlls)
│       │   └── *.dll
│       └── APPID_2
│           └── *.dll
│
│
├── app1                        - the app1 main/base folder
│   ├── hostfxr.dll;...         - dlls that can't be moved.
│   ├── nbloader.dll            - NBLoader(will be moved if use patch)
│   ├── app1.deps.json
│   ├── app1.dll
│   ├── app1.exe
│   ├── app1.runtimeconfig.json
│   └── ...
│
│
└── app2                        - the app2 main/base folder
    ├── hostfxr.dll;...
    ├── nbloader.dll
    ├── app2.deps.json
    ├── app2.dll
    ├── app2.exe
    ├── app2.runtimeconfig.json
    └── ...
```
