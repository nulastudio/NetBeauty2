using Avalonia;
using Avalonia.Controls;
using Avalonia.Markup.Xaml;

namespace XamlControlsGallery.Pages
{
    public class DateTimePage : UserControl
    {
        public DateTimePage()
        {
            this.InitializeComponent();
        }

        private void InitializeComponent()
        {
            AvaloniaXamlLoader.Load(this);
        }
    }
}
