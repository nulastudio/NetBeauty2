using System;
using System.Collections.ObjectModel;
using Avalonia;
using Avalonia.Controls;
using Avalonia.Markup.Xaml;
using Avalonia.Media;
using XamlControlsGallery.ViewModels;

namespace XamlControlsGallery.Pages
{
    public class LayoutPage : UserControl
    {
        public LayoutPage()
        {
            this.InitializeComponent();

            DataContext = new LayoutPageViewModel();
        }

        private void InitializeComponent()
        {
            AvaloniaXamlLoader.Load(this);
        }
    }
}
