using SixLabors.ImageSharp.Formats.Png;

using System;
using System.Collections.Generic;
using System.Globalization;
using System.Reflection;
using System.Resources;
using System.Runtime.InteropServices;
using System.Windows;

using Res = NetFxTest.Properties.Resources;

namespace NetFxTest
{
    /// <summary>
    /// Interaction logic for MainWindow.xaml
    /// </summary>
    public partial class MainWindow : Window
    {
        private static PngEncoder encoder;

        public MainWindow()
        {
            InitializeComponent();

            encoder = new PngEncoder();

            WindowStartupLocation = WindowStartupLocation.CenterScreen;
            RunTest();
        }

        private void RunTest()
        {
            bool resourcesResult = ResourcesTest();
            bool nativeResult = NativeDLLTest();

            bool testResult = resourcesResult && nativeResult;

            if (testResult)
            {
                TestResult.Content = "It Works!";
            }
            else
            {
                TestResult.Content = "Oops!";
            }
        }

        private bool ResourcesTest()
        {
            bool result = true;

            result &= InternalResourcesTest();
            result &= ExternalResourcesTest();

            return result;
        }

        private bool InternalResourcesTest()
        {
            var content = new Dictionary<string, string>()
            {
                { "en", "Hello"},
                { "zh-Hans", "你好"},
                { "zh-Hant", "你好"},
                { "ja", "こんにちは"},
                { "fr", "Bonjour"},
                { "de", "Hallo"},
                { "ko", "Hello(Default)"}, // not exist, should fallback to default "Hello"
            };

            foreach (var kv in content)
            {
                var culture = new CultureInfo(kv.Key);
                var text = Res.ResourceManager.GetString("Hello", culture);

                if (text != kv.Value) return false;
            }

            return true;
        }

        private bool ExternalResourcesTest()
        {
            if (encoder == null) return true;

            var content = new Dictionary<string, string>()
            {
                { "zh-Hans", "Memory<T> has been disposed."},
            };

            var rm = new ResourceManager("FxResources.System.Memory.SR", typeof(System.Buffers.Text.Base64).Assembly);

            foreach (var kv in content)
            {
                var culture = new CultureInfo(kv.Key);
                var text = rm.GetString("MemoryDisposed", culture);

                if (text != kv.Value) return false;
            }

            return true;
        }

        private bool NativeDLLTest()
        {
            bool result = true;

            result &= SumTest();

            return result;
        }

        private bool SumTest()
        {
            int a = 1;
            int b = 2;

            int result = a + b;
            int nativeResult = Util.Sum(a, b);

            return result == nativeResult;
        }
    }
}
