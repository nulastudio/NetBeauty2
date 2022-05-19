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

                var fileName = dllname;
                var assemblyPath = "";

                if (IsSharedRuntimeMode) {
                    assemblyPath = Path.GetFullPath($"{path}/srm_native/{SharedRuntimeAppID}/{fileName}");
                } else {
                    assemblyPath = Path.GetFullPath($"{path}/{fileName}");
                }

                if (File.Exists(assemblyPath))
                {
                    if (NativeLibrary.TryLoad(assemblyPath, out var handle)) {
                        return handle;
                    }
                }
            }

            return IntPtr.Zero;
        }
    }
}
