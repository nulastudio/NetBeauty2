using System;
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
        private static readonly string CK = "NetBeautyLibsDir";

        public static readonly string APP_BASE = AppContext.BaseDirectory ?? "";
        public static readonly string LIB_DIRECTORIES = AppContext.GetData(CK)?.ToString() ?? "";

        public static readonly string[] probes = LIB_DIRECTORIES.Split(';');

        public static Assembly ManagedAssemblyResolver(AssemblyLoadContext context, AssemblyName assemblyName)
        {
            foreach (var probe in probes)
            {
                if (probe == "") continue;

                var path = Path.IsPathRooted(probe) ? probe : $"{APP_BASE}/{probe}";

                var culture = assemblyName.CultureName ?? "";

                if (culture != "") {
                    culture = "/" + culture + "/";
                }

                var assemblyPath = Path.GetFullPath($"{path}{culture}{assemblyName.Name}.dll");

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

                var assemblyPath = Path.GetFullPath($"{path}/{dllname}");

                if (File.Exists(assemblyPath) && NativeLibrary.TryLoad(assemblyPath, out var handle))
                {
                    return handle;
                }
            }

            return IntPtr.Zero;
        }
    }
}
