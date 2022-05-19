using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;
using System.Windows.Documents;
using System.Windows.Input;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using System.Windows.Navigation;
using System.Windows.Shapes;

using WatsonWebsocket;

namespace WsClient
{
    /// <summary>
    /// Interaction logic for MainWindow.xaml
    /// </summary>
    public partial class MainWindow : Window
    {
        private WatsonWsClient client;

        public MainWindow()
        {
            InitializeComponent();
        }

        private async void Connect_Click(object sender, RoutedEventArgs e)
        {
            if (this.client != null && this.client.Connected)
            {
                this.client.Stop();
            }

            var part = Address.Text.Split(":");
            var ip = part[0];
            var port = int.Parse(part[1]);

            WatsonWsClient client = new WatsonWsClient(ip, port, false);
            client.ServerConnected += ServerConnected;
            client.ServerDisconnected += ServerDisconnected;
            client.MessageReceived += MessageReceived;
            await client.StartAsync();

            if (client.Connected)
            {
                this.client = client;
            }
            else
            {
                Content.Text += $"Can not connect to the server\n\n";
            }
        }

        private void DisConnect_Click(object sender, RoutedEventArgs e)
        {
            if (this.client == null || !this.client.Connected)
            {
                Content.Text += $"Server Unconnect\n\n";
                return;
            }

            this.client.Stop();
            this.client.Dispose();
            this.client = null;
        }

        private async void Send_Click(object sender, RoutedEventArgs e)
        {
            var text = SendText.Text;

            if (!string.IsNullOrEmpty(text))
            {
                if (this.client == null || !this.client.Connected)
                {
                    Content.Text += $"Server Unconnect\n\n";
                }
                else
                {
                    await this.client.SendAsync(text);
                    SendText.Text = "";
                    Content.Text += $"Send:\n{text}\n\n";
                }
            }

            SendText.Focus();
        }

        private void MessageReceived(object sender, MessageReceivedEventArgs args)
        {
            Application.Current.Dispatcher.Invoke(() =>
            {
                Content.Text += $"Receive:\n{Encoding.UTF8.GetString(args.Data)}\n\n";
            });
        }

        private void ServerConnected(object sender, EventArgs args)
        {
            Application.Current.Dispatcher.Invoke(() =>
            {
                Content.Text += $"Server Connect\n\n";
            });
        }

        private void ServerDisconnected(object sender, EventArgs args)
        {
            Application.Current.Dispatcher.Invoke(() =>
            {
                Content.Text += $"Server Disconnect\n\n";
            });
        }
    }
}
