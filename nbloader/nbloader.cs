using System;
using System.Collections.Generic;
using System.Runtime.Loader;
using System.Reflection;
using System.Runtime.InteropServices;
using System.IO;

namespace NetBeauty
{
    public class NBLoader
    {
        public readonly string APP_BASE;

        public readonly string LIB_DIRECTORIES;

        // "" "no" "default"
        public readonly string SharedRuntimeMode;
        public readonly string SharedRuntimeAppID;
        public readonly bool IsSharedRuntimeMode;
        public readonly Dictionary<string, string> srmMapping;

        public readonly string[] probes;

        public NBLoader()
        {
            APP_BASE = AppContext.BaseDirectory ?? "";
            LIB_DIRECTORIES = AppContext.GetData("NetBeautyLibsDir")?.ToString() ?? "";
            SharedRuntimeMode = AppContext.GetData("NetBeautySharedRuntimeMode")?.ToString() ?? "";
            SharedRuntimeAppID = AppContext.GetData("NetBeautyAppID")?.ToString() ?? "";
            IsSharedRuntimeMode = SharedRuntimeMode != "" && SharedRuntimeMode != "no";

            srmMapping = new Dictionary<string, string>();
            var mapping = AppContext.GetData("NetBeautySharedRuntimeMapping")?.ToString() ?? "";
            foreach (var v in mapping.Split('|'))
            {
                if (v == "") continue;

                var map = v.Split(':');

                if (map.Length != 2) continue;

                srmMapping[map[0]] = map[1];
            }

            probes = LIB_DIRECTORIES.Split(';');
        }

        public void RegisterALC(AssemblyLoadContext context)
        {
            context.Resolving += ManagedAssemblyResolver;
            context.ResolvingUnmanagedDll += UnmanagedDllResolver;
        }

        public Assembly ManagedAssemblyResolver(AssemblyLoadContext context, AssemblyName assemblyName)
        {
            foreach (var probe in probes)
            {
                if (string.IsNullOrEmpty(probe)) continue;

                var absPath = Path.IsPathRooted(probe) ? probe : $"{APP_BASE}/{probe}";

                var culture = assemblyName.CultureName ?? "";

                var culturePath = culture != "" ? $"locales/{culture}" : culture;

                var fileName = $"{assemblyName.Name}.dll";

                string assemblyPath;

                if (IsSharedRuntimeMode) {
                    var srmKey = culture != "" ? $"{culture}/{fileName}" : fileName;

                    var md5 = srmMapping.GetValueOrDefault(srmKey) ?? "";

                    assemblyPath = Path.GetFullPath($"{absPath}/{culturePath}/{fileName}/{md5}/{fileName}");
                } else {
                    assemblyPath = Path.GetFullPath($"{absPath}/{culturePath}/{fileName}");
                }

                if (File.Exists(assemblyPath))
                {
                    return context.LoadFromAssemblyPath(assemblyPath);
                }
            }

            return null;
        }

        public IntPtr UnmanagedDllResolver(Assembly assembly, string dllname)
        {
            foreach (var probe in probes)
            {
                if (string.IsNullOrEmpty(probe)) continue;

                var absPath = Path.IsPathRooted(probe) ? probe : $"{APP_BASE}/{probe}";

                foreach (var libname in DLLNameVariations(dllname))
                {
                    string assemblyPath;

                    if (IsSharedRuntimeMode) {
                        assemblyPath = Path.GetFullPath($"{absPath}/srm_native/{SharedRuntimeAppID}/{libname}");
                    } else {
                        assemblyPath = Path.GetFullPath($"{absPath}/{libname}");
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
