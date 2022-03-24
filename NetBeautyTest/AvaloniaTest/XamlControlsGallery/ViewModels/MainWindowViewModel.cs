using System.Collections.Generic;
using System.Reactive;
using ReactiveUI;

namespace XamlControlsGallery.ViewModels
{
    public class MainWindowViewModel : ViewModelBase
    {
        private bool _isMenuItemChecked;

        public MainWindowViewModel()
        {
            ToggleMenuItemCheckedCommand = ReactiveCommand.Create(() =>
            {
                IsMenuItemChecked = !IsMenuItemChecked;
            });
        }

        public bool IsMenuItemChecked
        {
            get { return _isMenuItemChecked; }
            set { this.RaiseAndSetIfChanged(ref _isMenuItemChecked, value); }
        }

        public ReactiveCommand<Unit, Unit> ToggleMenuItemCheckedCommand { get; }

        public string Greeting => "Welcome to Avalonia!";

        public List<string> SearchItems { get; } = new List<string>
        {
            "TextBlock",
            "CheckBox",
            "ComboBox",
            "TextBox",
            "Calendar"
        };
    }
}
