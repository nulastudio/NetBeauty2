using System;
using System.Collections.ObjectModel;
using Avalonia.Media;

namespace XamlControlsGallery.ViewModels
{
    public class ItemViewModel
    {
        public ItemViewModel(int i, Color color)
        {
            Text = i.ToString();
            Color = color;
        }

        public string Text { get; }

        public Color Color { get; }
    }

    public class LayoutPageViewModel : ViewModelBase
    {
        public LayoutPageViewModel()
        {
            Items = new ObservableCollection<ItemViewModel>();

            var availableColors = new[]
            {
                Colors.Red, Colors.Blue, Colors.CornflowerBlue, Colors.Gray, Colors.Magenta, Colors.Brown, Colors.Orange, Colors.Crimson
            };

            var random = new Random(0);

            for (int i = 0; i < 10000; i++)
            {
                var color = availableColors[random.Next(0, availableColors.Length)];

                var item = new ItemViewModel(i, color);

                Items.Add(item);
            }
        }

        public ObservableCollection<ItemViewModel> Items { get; }
    }
}
