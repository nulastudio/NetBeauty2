using System;
using System.Collections.Generic;
using System.Configuration;
using System.Globalization;
using System.Reflection;
using System.Resources;
using System.Runtime.InteropServices;
using System.Runtime.Loader;
using System.Windows;

using Res = WPFTest.Properties.Resources;

namespace WPFTest
{
    /// <summary>
    /// Interaction logic for MainWindow.xaml
    /// </summary>
    public partial class MainWindow : Window
    {
        public MainWindow()
        {
            InitializeComponent();
            
            WindowStartupLocation = WindowStartupLocation.CenterScreen;
            RunTest();
        }

        private void RunTest()
        {
            bool resourcesResult = ResourcesTest();
            bool nativeResult = NativeDLLTest();
            bool configResult = ConfigurationTest();

            bool testResult = resourcesResult && nativeResult && configResult;

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
            var content = new Dictionary<string, string>()
            {
                { "zh-Hans", "帮助"},
                { "de", "Hilfe"},
            };

            var rm = new ResourceManager("FxResources.PresentationCore.SR", typeof(System.Windows.UIElement).Assembly);

            foreach (var kv in content)
            {
                var culture = new CultureInfo(kv.Key);
                var text = rm.GetString("HelpText", culture);

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

        private bool ConfigurationTest()
        {
            string config = ConfigurationManager.AppSettings["hello"];

            return config == "你好";
        }
    }
}
