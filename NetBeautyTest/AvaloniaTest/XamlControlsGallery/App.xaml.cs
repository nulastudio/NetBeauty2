using Avalonia;
using Avalonia.Controls;
using Avalonia.Controls.ApplicationLifetimes;
using Avalonia.Markup.Xaml;
using XamlControlsGallery.ViewModels;
using XamlControlsGallery.Views;

namespace XamlControlsGallery
{
    public class App : Application
    {
        public override void Initialize()
        {
            AvaloniaXamlLoader.Load(this);
        }

        public override void OnFrameworkInitializationCompleted()
        {
            if (ApplicationLifetime is IClassicDesktopStyleApplicationLifetime desktop)
            {
                desktop.MainWindow = new MainWindow
                {
                    DataContext = new MainWindowViewModel(),
                };
            }
            else if (ApplicationLifetime is ISingleViewApplicationLifetime singleViewLifetime)
            {
                singleViewLifetime.MainView = new MainView
                {
                    DataContext = new MainWindowViewModel(),
                };
            }

            var theme = new Avalonia.Themes.Default.DefaultTheme();
            theme.TryGetResource("Button", out _);

            //var theme1 = new Avalonia.Themes.Fluent.FluentTheme();
            //theme1.TryGetResource("Button", out _);

            base.OnFrameworkInitializationCompleted();
        }
    }
}
