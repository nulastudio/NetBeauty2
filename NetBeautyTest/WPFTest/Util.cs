using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading.Tasks;

namespace WPFTest
{
    class Util
    {
        public static readonly bool Is64Bit = (IntPtr.Size == 8);

        [DllImport("libraries/sum/x86/sum", CallingConvention = CallingConvention.Cdecl, EntryPoint = "sum")]
        private static extern int sum_x86(int a, int b);

        [DllImport("libraries/sum/x64/sum", CallingConvention = CallingConvention.Cdecl, EntryPoint = "sum")]
        private static extern int sum_x64(int a, int b);

        public static int Sum(int a, int b)
        {
            return Is64Bit ? sum_x64(a, b) : sum_x86(a, b);
        }
    }
}
