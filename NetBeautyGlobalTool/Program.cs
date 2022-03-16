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
            ["win-x86"]   = "/nbeauty/win-x86/nbeauty2.exe",
            ["win-x64"]   = "/nbeauty/win-x64/nbeauty2.exe",
            ["linux-x64"] = "/nbeauty/linux-x64/nbeauty2",
            ["osx-x64"]   = "/nbeauty/osx-x64/nbeauty2",
        };

        static void Main(string[] args)
        {
            var rootDir = Path.GetDirectoryName(Assembly.GetEntryAssembly().Location) + "/../..";
            var nbeautyBin = "";
            if (RuntimeInformation.IsOSPlatform(OSPlatform.Windows))
            {
                nbeautyBin = platform["win-x86"];
            }
            else if (RuntimeInformation.IsOSPlatform(OSPlatform.Linux))
            {
                nbeautyBin = platform["linux-x64"];
            }
            else if (RuntimeInformation.IsOSPlatform(OSPlatform.OSX))
            {
                nbeautyBin = platform["osx-x64"];
            }
            var psi = new ProcessStartInfo(rootDir + nbeautyBin)
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
