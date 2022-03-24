using Avalonia;
using Avalonia.Controls;
using Avalonia.Markup.Xaml;

namespace XamlControlsGallery.Pages
{
    public class OverviewPage : UserControl
    {
        public OverviewPage()
        {
            this.InitializeComponent();
        }

        private void InitializeComponent()
        {
            AvaloniaXamlLoader.Load(this);
        }
    }
}
