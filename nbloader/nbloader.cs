using System;
using System.Collections.Generic;
using System.Runtime.Loader;
using System.Reflection;
using System.Runtime.InteropServices;
using System.IO;

internal class StartupHook
{
    public static void Initialize()
    {
        AssemblyLoadContext.Default.Resolving += NetBeauty.NBLoader.ManagedAssemblyResolver;
        AssemblyLoadContext.Default.ResolvingUnmanagedDll += NetBeauty.NBLoader.UnmanagedDllResolver;
    }
}

namespace NetBeauty
{
    internal static class NBLoader
    {
        public static readonly string APP_BASE = AppContext.BaseDirectory ?? "";
        public static readonly string LIB_DIRECTORIES = AppContext.GetData("NetBeautyLibsDir")?.ToString() ?? "";
        // modes: ""(equals "no"), "no", "default"
        public static readonly string SharedRuntimeMode = AppContext.GetData("NetBeautySharedRuntimeMode")?.ToString() ?? "";
        public static readonly string SharedRuntimeAppID = AppContext.GetData("NetBeautyAppID")?.ToString() ?? "";
        public static readonly bool IsSharedRuntimeMode = SharedRuntimeMode != "" && SharedRuntimeMode != "no";
        private static Dictionary<string, string> _srmMapping;
        public static Dictionary<string, string> srmMapping
        {
            get
            {
                if (_srmMapping == null) {
                    _srmMapping = new Dictionary<string, string>();
                    var mapping = AppContext.GetData("NetBeautySharedRuntimeMapping")?.ToString() ?? "";
                    foreach (var v in mapping.Split('|'))
                    {
                        var map = v.Split(':');

                        _srmMapping[map[0]] = map[1];
                    }
                }

                return _srmMapping;
            }
        }

        public static readonly string[] probes = LIB_DIRECTORIES.Split(';');

        public static Assembly ManagedAssemblyResolver(AssemblyLoadContext context, AssemblyName assemblyName)
        {
            foreach (var probe in probes)
            {
                if (probe == "") continue;

                var path = Path.IsPathRooted(probe) ? probe : $"{APP_BASE}/{probe}";

                var culture = assemblyName.CultureName ?? "";

                var culturePath = culture;
                if (culturePath != "") {
                    culturePath = "locales/" + culturePath;
                }

                var fileName = $"{assemblyName.Name}.dll";
                var assemblyPath = "";

                if (IsSharedRuntimeMode) {
                    var srmKey = fileName;
                    if (culture != "") {
                        srmKey = $"{culture}/{srmKey}";
                    }
                    var md5 = srmMapping.GetValueOrDefault(srmKey);
                    if (md5 == null) md5 = "";

                    assemblyPath = Path.GetFullPath($"{path}/{culturePath}/{fileName}/{md5}/{fileName}");
                } else {
                    assemblyPath = Path.GetFullPath($"{path}/{culturePath}/{fileName}");
                }

                if (File.Exists(assemblyPath))
                {
                    return context.LoadFromAssemblyPath(assemblyPath);
                }
            }

            return null;
        }

        public static IntPtr UnmanagedDllResolver(Assembly assembly, string dllname)
        {
            foreach (var probe in probes)
            {
                if (probe == "") continue;

                var path = Path.IsPathRooted(probe) ? probe : $"{APP_BASE}/{probe}";

                foreach (var libname in DLLNameVariations(dllname))
                {
                    var assemblyPath = "";

                    if (IsSharedRuntimeMode) {
                        assemblyPath = Path.GetFullPath($"{path}/srm_native/{SharedRuntimeAppID}/{libname}");
                    } else {
                        assemblyPath = Path.GetFullPath($"{path}/{libname}");
                    }

                    if (File.Exists(assemblyPath))
                    {
                        if (NativeLibrary.TryLoad(assemblyPath, out var handle)) {
                            return handle;
                        }
                    }
                }
            }

            return IntPtr.Zero;
        }

        // 参考：https://docs.microsoft.com/zh-cn/dotnet/standard/native-interop/cross-platform#library-name-variations
        public static string[] DLLNameVariations(string dllname)
        {
            var names = new List<string>();

            var isWindows = RuntimeInformation.IsOSPlatform(OSPlatform.Windows);
            var isOSX = RuntimeInformation.IsOSPlatform(OSPlatform.OSX);

            if (isWindows)
            {
                names.Add(dllname);

                if (!dllname.EndsWith(".dll") && !dllname.EndsWith(".exe")) {
                    names.Add(dllname + ".dll");
                    names.Add(dllname + ".exe");
                }
            }
            else if (isOSX)
            {
                names.Add(dllname);
                names.Add("lib" + dllname);

                if (!dllname.EndsWith(".dylib")) {
                    names.Add(dllname + ".dylib");
                    names.Add("lib" + dllname + ".dylib");
                }
            }
            else
            {
                names.Add(dllname);
                names.Add("lib" + dllname);

                if (!dllname.EndsWith(".so")) {
                    names.Add(dllname + ".so");
                    names.Add("lib" + dllname + ".so");
                }
            }

            return names.ToArray();
        }
    }
}
