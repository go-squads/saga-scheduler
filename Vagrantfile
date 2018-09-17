# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure("2") do |config|
  config.vm.define "lxd1" do |lxd1|
    lxd1.vm.hostname = "lxd1"
    lxd1.vm.box = "ubuntu/xenial64"

    lxd1.vm.network "private_network", type: "dhcp"
    #vm1.vm.synced_folder "/Users/pt.go-jekindonesia/lxd1", "/home"
    lxd1.vm.provider "virtualbox" do |vb|
      vb.name = "lxd1"
      # Display the VirtualBox GUI when booting the machine
      vb.gui = false
      # Customize the amount of memory on the VM:
      vb.memory = "512"
    end
  end

  config.vm.define "lxd2" do |lxd2|
    lxd2.vm.hostname = "lxd2"
    lxd2.vm.box = "ubuntu/xenial64"

    lxd2.vm.network "private_network", type: "dhcp"
    #vm2.vm.synced_folder "/Users/pt.go-jekindonesia/lxd2", "/home"
    lxd2.vm.provider "virtualbox" do |vb|
      vb.name = "lxd2"
      # Display the VirtualBox GUI when booting the machine
      vb.gui = false
      # Customize the amount of memory on the VM:
      vb.memory = "512"
    end
  end

  config.vm.define "scheduler" do |scheduler|
    scheduler.vm.hostname = "scheduler"
    scheduler.vm.box = "ubuntu/xenial64"

    scheduler.vm.network "private_network", type: "dhcp"
    scheduler.vm.network "forwarded_port", host: 5433, guest: 5432
    scheduler.vm.provider "virtualbox" do |vb|
      vb.name = "metric-collector"
      # Display the VirtualBox GUI when booting the machine
      vb.gui = false
      # Customize the amount of memory on the VM:
      vb.memory = "1024"
    end
  end

end
