using System;
using System.IO;
using System.Reflection;
using System.Collections;
using System.Diagnostics;
using System.Linq;
using System.Collections.Generic;
using System.Runtime.InteropServices;

namespace NetCoreBeautyGlobalTool
{
    class Program
    {
        private static readonly Dictionary<string, string> platform = new Dictionary<string, string> {
            ["win-x86"]   = "/ncbeauty/win-x86/ncbeauty.exe",
            ["win-x64"]   = "/ncbeauty/win-x64/ncbeauty.exe",
            ["linux-x64"] = "/ncbeauty/linux-x64/ncbeauty",
            ["osx-x64"]   = "/ncbeauty/osx-x64/ncbeauty",
        };

        static void Main(string[] args)
        {
            var rootDir = Path.GetDirectoryName(Assembly.GetEntryAssembly().Location) + "/../..";
            var ncbeautyBin = "";
            if (RuntimeInformation.IsOSPlatform(OSPlatform.Windows))
            {
                ncbeautyBin = platform["win-x86"];
            }
            else if (RuntimeInformation.IsOSPlatform(OSPlatform.Linux))
            {
                ncbeautyBin = platform["linux-x64"];
            }
            else if (RuntimeInformation.IsOSPlatform(OSPlatform.OSX))
            {
                ncbeautyBin = platform["osx-x64"];
            }
            var psi = new ProcessStartInfo(rootDir + ncbeautyBin)
            {
                UseShellExecute        = false,
                CreateNoWindow         = true,
                RedirectStandardOutput = true,
                RedirectStandardError  = true,
            };

            foreach (var arg in args)
                psi.ArgumentList.Add(arg);

            using (var process = Process.Start(psi))
            {
                process.OutputDataReceived += (_, ea) =>
                {
                    if (ea.Data != null)
                        Console.WriteLine(ea.Data);
                };

                process.ErrorDataReceived += (_, ea) =>
                {
                    if (ea.Data != null)
                        Console.Error.WriteLine(ea.Data);
                };

                process.BeginOutputReadLine();
                process.BeginErrorReadLine();

                process.WaitForExit();

                Environment.Exit(process.ExitCode);
            }
        }
    }
}
