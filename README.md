# NetBeauty 2

## Overview

**NetBeauty 2** is a tool designed to organize your .NET Framework or .NET Core application's runtime components and dependencies into a sub-directory, resulting in a cleaner and more manageable project structure.

## Key Features

- Supports both .NET Framework and .NET Core 3.0+;
- Compatible with all platforms;
- Works with various deployment models: Framework-dependent (FDD), Self-contained (SCD), and Framework-dependent executables (FDE);
- Reduces file clutter by moving dependencies into a dedicated directory;
- Offers advanced options for hiding unnecessary files and customizing the runtime structure.

## Visual Comparison

### Before Applying NetBeauty

![Before Beauty](screenshot/before_beauty.png)

### After Applying NetBeauty

![After Beauty](screenshot/after_beauty.png)

**Even fewer files!**  

Explore the [`--hiddens`](#using-the-binary-application-for-published-projects) option for further reduction.

![After Beauty with Hiddens](screenshot/after_beauty_with_hiddens.png)

## What's New in NetBeauty 2?

### NetBeauty 2 vs NetCoreBeauty

#### Supported Frameworks

- **NetBeauty 2:** .NET Framework, .NET Core 3.0+
- **NetCoreBeauty:** .NET Core 2.0+

#### Deployment Models

- **NetBeauty 2:** FDD, SCD, FDE
- **NetCoreBeauty:** SCD only

#### Supported Platforms

- **NetBeauty 2:** All platforms
- **NetCoreBeauty:** Windows (x64, x86, arm64), Linux (x64, arm, arm64), macOS (x64, arm64)

#### Patched HostFXR Requirement

- **NetBeauty 2:** Not required (except when using patch)
- **NetCoreBeauty:** Required

#### Minimum File Structure

- **NetBeauty 2:** ~20 files (default), ~8 files (with patch)
- **NetCoreBeauty:** ~8 files

#### How It Works

- **NetBeauty 2:**
  - Utilizes [`STARTUP_HOOKS`](https://github.com/dotnet/runtime/blob/main/docs/design/features/host-startup-hook.md)
  - Handles assembly resolution via [`AssemblyLoadContext.Resolving`](https://docs.microsoft.com/en-us/dotnet/api/system.runtime.loader.assemblyloadcontext.resolving?view=netcore-3.0) and [`AssemblyLoadContext.ResolvingUnmanagedDll`](https://docs.microsoft.com/en-us/dotnet/api/system.runtime.loader.assemblyloadcontext.resolvingunmanageddll?view=netcore-3.0)
  - Optionally uses a [patched libhostfxr](https://github.com/nulastudio/HostFXRPatcher) and [`additionalProbingPaths`](https://github.com/dotnet/toolset/blob/master/Documentation/specs/runtime-configuration-file.md#runtimeoptions-section-runtimeconfigjson) when patching
- **NetCoreBeauty:**
  - Relies on [patched libhostfxr](https://github.com/nulastudio/HostFXRPatcher)
  - Uses [`additionalProbingPaths`](https://github.com/dotnet/toolset/blob/master/Documentation/specs/runtime-configuration-file.md#runtimeoptions-section-runtimeconfigjson)

#### Shared Runtime Support

- **NetBeauty 2:** Yes
- **NetCoreBeauty:** Possible (with patched libhostfxr)

> [!TIP]  
> For more details, visit the [NetBeauty 2 repository](https://github.com/nulastudio/NetBeauty2) and the [NetCoreBeauty (v1) repository](https://github.com/nulastudio/NetBeauty2/tree/v1).

## The Patch Returns

One of the main objectives of NetBeauty 2 is to use a custom loader instead of patching. However, the loader requires many types and APIs (such as `Dictionary<TKey, TValue>`, `List<T>`, `Path.GetFullPath`, `File.Exists`, `NativeLibrary`, `RuntimeInformation`, etc.), which introduces numerous assembly references. Unfortunately, these files cannot be moved; otherwise, CoreCLR fails to initialize and invoke the loader. More complex logic leads to more files. Therefore, the patch is still necessary.

Now, both the loader and the patch work seamlessly together:

- The loader enables support for FDD and FDE applications.
- The patch minimizes the file count (SCD apps only).

## Breaking Changes in v2.1.5

The startup hook has been renamed from `nbloader` to `libloader`.  
No action is required on your part. Both `BeautyNBLoaderVerPolicy` (in the project file) and `nbloaderverpolicy` (in the CLI) remain unchanged for maximum backward compatibility. For more information, see [issue #80](https://github.com/nulastudio/NetBeauty2/issues/80).

## Getting Started

### Adding NetBeauty via NuGet

To add NetBeauty to your .NET Core project, run:

```bash
dotnet add package nulastudio.NetBeauty
```

Your `.csproj` file should look similar to the following:

```xml
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>netcoreapp3.0</TargetFramework>
  </PropertyGroup>

  <PropertyGroup>
    <BeautySharedRuntimeMode>False</BeautySharedRuntimeMode>
    <!-- Move dependencies into a sub-directory (default: libs). Use quotes if the path contains spaces. -->
    <BeautyLibsDir Condition="$(BeautySharedRuntimeMode) == 'True'">../libraries</BeautyLibsDir>
    <BeautyLibsDir Condition="$(BeautySharedRuntimeMode) != 'True'">./libraries</BeautyLibsDir>
    <!-- Exclude specific DLLs from being moved -->
    <!-- <BeautyExcludes>dll1.dll;lib*;...</BeautyExcludes> -->
    <!-- Hide files not needed by end users -->
    <!-- <BeautyHiddens>hostfxr;hostpolicy;*.deps.json;*.runtimeconfig*.json</BeautyHiddens> -->
    <DisableBeauty>False</DisableBeauty>
    <BeautyOnPublishOnly>False</BeautyOnPublishOnly>
    <BeautyNoRuntimeInfo>False</BeautyNoRuntimeInfo>
    <BeautyNBLoaderVerPolicy>auto</BeautyNBLoaderVerPolicy>
    <BeautyEnableDebugging>False</BeautyEnableDebugging>
    <BeautyUsePatch>True</BeautyUsePatch>
    <!-- Customize AppHost entry and directory if needed -->
    <!-- <BeautyAppHostEntry>bin/MyApp.dll</BeautyAppHostEntry> -->
    <!-- <BeautyAppHostDir>..</BeautyAppHostDir> -->
    <!-- <BeautyAfterTasks></BeautyAfterTasks> -->
    <BeautyLogLevel>Info</BeautyLogLevel>
    <!-- Set a repository mirror if you have trouble connecting to GitHub -->
    <!-- <BeautyGitCDN>https://gitee.com/liesauer/HostFXRPatcher</BeautyGitCDN> -->
    <!-- <BeautyGitTree>master</BeautyGitTree> -->
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="nulastudio.NetBeauty" Version="2.1.5.0" />
  </ItemGroup>
</Project>
```

After configuring your project, simply run `dotnet build` or `dotnet publish`. NetBeauty will handle everything automatically.

### Using the Binary Application for Published Projects

If your project is already published, you can use the NetBeauty binary application:

```bash
# Usage:
nbeauty2 [--loglevel=(Error|Detail|Info)] [--srmode] [--enabledebug] [--usepatch] [--hiddens=hiddenFiles] [--noruntimeinfo] [--roll-forward=<rollForward>] [--nbloaderverpolicy=(auto|with|without)] [--apphostentry=<appHostEntry>] [--apphostdir=<appHostDir>] <beautyDir> [<libsDir> [<excludes>]]
```

**Example:**

```bash
nbeauty2 --usepatch --loglevel Detail --hiddens "hostfxr;hostpolicy;*.deps.json;*.runtimeconfig*.json" "/path/to/publishDir" libraries "dll1.dll;lib*;..."
```

> [!NOTE]  
> The `--hiddens` option only hides files (does not move them) and is supported on Windows only.

### Installing as a .NET Core Global Tool

To install NetBeauty as a global tool, run:

```bash
dotnet tool install --global nulastudio.nbeauty
```

You can then use it as you would any other binary distribution.

## Shared Runtime Structure

Below is an example of a shared runtime directory structure:

```bash
├── libraries                   # Shared runtime DLLs (customizable name)
│   ├── locales                 # Satellite assemblies
│   │   ├── en
│   │   │   └── *.resources.dll
│   │   │       ├── MD5_1       # Allows multiple runtimes between apps
│   │   │       │   └── *.resources.dll
│   │   │       └── MD5_2
│   │   │           └── *.resources.dll
│   │   ├── zh-Hans
│   │   │   └── *.resources.dll
│   │   │       ├── MD5_1
│   │   │       │   └── *.resources.dll
│   │   │       └── MD5_2
│   │   │           └── *.resources.dll
│   │   └── ...                 # Other languages
│   ├── *.dll                   # Shared managed assemblies
│   │   ├── MD5_1
│   │   │   └── *.dll
│   │   └── MD5_2
│   │       └── *.dll
│   └── srm_native              # Native DLLs (not shared; each app has its own copy)
│       ├── APPID_1
│       │   └── *.dll
│       └── APPID_2
│           └── *.dll
├── app1                        # Main folder for app1
│   ├── hostfxr.dll ...         # DLLs that cannot be moved
│   ├── libloader.dll           # Loader (moved if using patch)
│   ├── app1.deps.json
│   ├── app1.dll
│   ├── app1.exe
│   ├── app1.runtimeconfig.json
│   └── ...
└── app2                        # Main folder for app2
    ├── hostfxr.dll ...
    ├── libloader.dll
    ├── app2.deps.json
    ├── app2.dll
    ├── app2.exe
    ├── app2.runtimeconfig.json
    └── ...
```

## Customizing AppHost

NetBeauty 2 draws inspiration from [AppHostPatcher](https://github.com/dnSpy/dnSpy/tree/master/Build/AppHostPatcher) to provide a more user-friendly folder structure for software suites by patching the imprinted entry path of AppHost.  
See the [demo](https://github.com/nulastudio/NetBeauty2/tree/master/NetBeautyTest/SharedRuntimeTest) for more details.

**Example Structure:**

```bash
├── MyApp                       # Main folder for the app
│   ├── libs                    # Dependencies
│   ├── hostfxr.dll ...         # DLLs that cannot be moved
│   ├── libloader.dll           # Loader (moved if using patch)
│   ├── MyApp.deps.json
│   ├── MyApp.dll
│   ├── MyApp.runtimeconfig.json
│   └── ...
└── MyApp.exe                   # AppHost
```

**Shared Runtime with Customized AppHost:**

```bash
├── libraries                   # Shared runtime DLLs (customizable name)
├── app1                        # Main folder for app1
│   ├── hostfxr.dll ...
│   ├── app1.deps.json
│   ├── app1.dll
│   ├── app1.runtimeconfig.json
│   └── ...
├── app2                        # Main folder for app2
│   ├── hostfxr.dll ...
│   ├── app2.deps.json
│   ├── app2.dll
│   ├── app2.runtimeconfig.json
│   └── ...
├── app1.exe
└── app2.exe
```

## License

This project is licensed under the MIT License.

```txt
Copyright © 2022 nullastudio

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```
