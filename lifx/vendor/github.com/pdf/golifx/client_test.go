package golifx_test

import (
	"errors"
	"time"

	. "github.com/pdf/golifx"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"

	"github.com/pdf/golifx/common"
	"github.com/pdf/golifx/mocks"
	"github.com/stretchr/testify/mock"
)

func init() {
	format.UseStringerRepresentation = false
}

var _ = Describe("Golifx", func() {
	var (
		client               *Client
		protocolSubscription *common.Subscription
		clientSubscription   *common.Subscription
		timeout              = 500 * time.Millisecond

		mockProtocol *mocks.Protocol
		mockDevice   *mocks.Device
		mockLight    *mocks.Light
		mockLocation *mocks.Location
		mockGroup    *mocks.Group

		deviceID           = uint64(1234)
		deviceUnknownID    = uint64(4321)
		deviceLabel        = `mockDevice`
		deviceUnknownLabel = `unknownDevice`
		lightID            = uint64(5678)
		lightLabel         = `mockLight`

		locationID           = `mockLocationID`
		locationUnknownID    = `unknownLocationID`
		locationLabel        = `mockLocation`
		locationUnknownLabel = `unknownLocation`
		groupID              = `mockGroupID`
		groupUnknownID       = `unknownGroupID`
		groupLabel           = `mockGroup`
		groupUnknownLabel    = `unknownGroup`
	)

	It("should send discovery to the protocol on NewClient", func() {
		var err error
		mockProtocol = new(mocks.Protocol)
		mockProtocol.On(`SetTimeout`, mock.AnythingOfType("*time.Duration")).Return().Once()
		mockProtocol.On(`SetRetryInterval`, mock.AnythingOfType("*time.Duration")).Return().Once()
		mockProtocol.On(`SetClient`, mock.Anything).Return().Once()
		mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(common.NewSubscription(mockProtocol), nil).Once()
		mockProtocol.On(`Discover`).Return(nil).Once()

		client, err = NewClient(mockProtocol)
		Expect(client).To(BeAssignableToTypeOf(new(Client)))
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Client", func() {
		BeforeEach(func() {
			mockProtocol = new(mocks.Protocol)
			protocolSubscription = common.NewSubscription(mockProtocol)
			mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
			mockProtocol.On(`SetClient`, mock.Anything).Return().Once()
			mockProtocol.On(`Discover`).Return(nil).Once()
			mockProtocol.On(`SetTimeout`, mock.AnythingOfType("*time.Duration")).Return().Once()
			mockProtocol.On(`SetRetryInterval`, mock.AnythingOfType("*time.Duration")).Return().Once()
			client, _ = NewClient(mockProtocol)
			client.SetTimeout(timeout)
			clientSubscription, _ = client.NewSubscription()

			mockDevice = new(mocks.Device)
			mockLight = new(mocks.Light)
			mockLocation = new(mocks.Location)
			mockGroup = new(mocks.Group)
		})

		AfterEach(func() {
			mockProtocol.SubscriptionTarget.On(`CloseSubscription`, mock.Anything).Return(nil).Once()
			mockProtocol.On(`Close`).Return(nil).Once()
			_ = client.Close()
		})

		It("should update the timeout", func() {
			t := 5 * time.Second
			mockProtocol.On(`SetTimeout`, t).Return().Once()
			client.SetTimeout(t)
			Expect(client.GetTimeout()).To(Equal(&t))
		})

		It("should update the retry interval", func() {
			interval := 5 * time.Millisecond
			client.SetRetryInterval(interval)
			Expect(client.GetRetryInterval()).To(Equal(&interval))
		})

		It("should set the retry to half the timeout if it's >= the timeout", func() {
			timeout := 10 * time.Second
			halfTimeout := timeout / 2
			mockProtocol.On(`SetTimeout`, mock.AnythingOfType("*time.Duration")).Return().Once()
			client.SetTimeout(timeout)
			interval := 10 * time.Second
			client.SetRetryInterval(interval)
			Expect(client.GetRetryInterval()).To(Equal(&halfTimeout))
		})

		It("should update the discovery interval", func() {
			interval := 5 * time.Second
			Expect(client.SetDiscoveryInterval(interval)).To(Succeed())
		})

		It("should update the discovery interval when it's non-zero", func() {
			interval := 5 * time.Second
			Expect(client.SetDiscoveryInterval(interval)).To(Succeed())
			interval = 10 * time.Second
			Expect(client.SetDiscoveryInterval(interval)).To(Succeed())
		})

		It("should perform discovery on the interval", func() {
			mockProtocol.On(`Discover`).Return(nil).Twice()
			Expect(client.SetDiscoveryInterval(100 * time.Millisecond)).To(Succeed())
			time.Sleep(250 * time.Millisecond)
			mockProtocol.AssertNumberOfCalls(GinkgoT(), `Discover`, 3)
		})

		It("should send SetPower to the protocol", func() {
			mockProtocol.On(`SetPower`, true).Return(nil).Once()
			Expect(client.SetPower(true)).To(Succeed())
		})

		It("should send SetPowerDuration to the protocol", func() {
			duration := 5 * time.Second
			mockProtocol.On(`SetPowerDuration`, true, duration).Return(nil).Once()
			Expect(client.SetPowerDuration(true, duration)).To(Succeed())
		})

		It("should send SetColor to the protocol", func() {
			color := common.Color{}
			duration := 1 * time.Millisecond
			mockProtocol.On(`SetColor`, color, duration).Return(nil).Once()
			Expect(client.SetColor(color, duration)).To(Succeed())
		})

		It("should return locations", func() {
			mockProtocol.On(`GetLocations`).Return([]common.Location{mockLocation}, nil).Once()
			locations, err := client.GetLocations()
			Expect(len(locations)).To(Equal(1))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return an error when it knows no locations", func() {
			mockProtocol.On(`GetLocations`).Return(nil, common.ErrNotFound).Once()
			locations, err := client.GetLocations()
			Expect(len(locations)).To(Equal(0))
			Expect(err).To(Equal(common.ErrNotFound))
		})

		It("should return groups", func() {
			mockProtocol.On(`GetGroups`).Return([]common.Group{mockGroup}, nil).Once()
			groups, err := client.GetGroups()
			Expect(len(groups)).To(Equal(1))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return an error when it knows no groups", func() {
			mockProtocol.On(`GetGroups`).Return(nil, common.ErrNotFound).Once()
			groups, err := client.GetGroups()
			Expect(len(groups)).To(Equal(0))
			Expect(err).To(Equal(common.ErrNotFound))
		})

		It("should return an error from GetDevices when it knows no devices", func() {
			mockProtocol.On(`GetDevices`).Return(nil, common.ErrNotFound).Once()
			devices, err := client.GetDevices()
			Expect(len(devices)).To(Equal(0))
			Expect(err).To(Equal(common.ErrNotFound))
		})

		It("should close successfully", func() {
			mockProtocol.On(`Close`).Return(nil).Once()
			Expect(client.Close()).To(Succeed())
		})

		It("should return an error on failed close", func() {
			mockProtocol.On(`Close`).Return(errors.New(`close failure`)).Once()
			Expect(client.Close()).NotTo(Succeed())
		})

		It("should return an error on double-close", func() {
			mockProtocol.On(`Close`).Return(nil).Once()
			Expect(client.Close()).To(Succeed())
			Expect(client.Close()).To(Equal(common.ErrClosed))
		})

		It("should publish an EventNewLocation on discovering a location", func(done Done) {
			mockLocation.Group.On(`ID`).Return(locationID).Once()
			event := common.EventNewLocation{Location: mockLocation}
			ch := make(chan interface{})
			go func() {
				evt := <-clientSubscription.Events()
				ch <- evt
			}()
			_ = protocolSubscription.Write(event)
			Expect(<-ch).To(Equal(event))
			close(done)
		})

		It("should publish an EventNewGroup on discovering a group", func(done Done) {
			mockGroup.On(`ID`).Return(groupID).Once()
			event := common.EventNewGroup{Group: mockGroup}
			ch := make(chan interface{})
			go func() {
				evt := <-clientSubscription.Events()
				ch <- evt
			}()
			_ = protocolSubscription.Write(event)
			Expect(<-ch).To(Equal(event))
			close(done)
		})

		It("should publish an EventNewDevice on discovering a device", func(done Done) {
			mockDevice.On(`ID`).Return(deviceID).Once()
			event := common.EventNewDevice{Device: mockDevice}
			ch := make(chan interface{})
			go func() {
				evt := <-clientSubscription.Events()
				ch <- evt
			}()
			_ = protocolSubscription.Write(event)
			Expect(<-ch).To(Equal(event))
			close(done)
		})

		Context("with locations", func() {

			Context("finding a location", func() {
				It("should find it by ID", func() {
					mockProtocol.On(`GetLocation`, locationID).Return(mockLocation, nil).Once()
					loc, err := client.GetLocationByID(locationID)
					Expect(loc).To(Equal(mockLocation))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return an error when the ID is not known", func() {
					mockProtocol.On(`GetLocation`, locationUnknownID).Return(&mocks.Location{}, common.ErrNotFound).Once()
					mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
					mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
					_, err := client.GetLocationByID(locationUnknownID)
					Expect(err).To(MatchError(common.ErrNotFound))
				})

				It("should find it by label", func() {
					mockProtocol.On(`GetLocations`).Return([]common.Location{mockLocation}, nil).Once()
					mockLocation.Group.On(`GetLabel`).Return(locationLabel, nil).Once()
					loc, err := client.GetLocationByLabel(locationLabel)
					Expect(loc).To(Equal(mockLocation))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return an error when the label is not known", func() {
					mockProtocol.On(`GetLocations`).Return([]common.Location{mockLocation}, nil).Once()
					mockLocation.Group.On(`GetLabel`).Return(locationLabel, nil).Once()
					mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
					mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
					_, err := client.GetLocationByLabel(locationUnknownLabel)
					Expect(err).To(MatchError(common.ErrNotFound))
				})

				Context("when the location is added while searching", func() {

					It("should find it by ID", func(done Done) {
						locChan := make(chan common.Location)
						errChan := make(chan error)
						mockProtocol.On(`GetLocation`, locationUnknownID).Return(&mocks.Location{}, common.ErrNotFound).Once()
						protocolSubscription = common.NewSubscription(mockProtocol)
						mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
						mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
						go func() {
							loc, err := client.GetLocationByID(locationUnknownID)
							errChan <- err
							locChan <- loc
						}()
						unknownLocation := new(mocks.Location)
						unknownLocation.Group.On(`ID`).Return(locationUnknownID).Once()
						_ = protocolSubscription.Write(common.EventNewLocation{Location: unknownLocation})
						Expect(<-errChan).NotTo(HaveOccurred())
						Expect(<-locChan).To(Equal(unknownLocation))
						close(done)
					})

					It("should find it by label", func(done Done) {
						locChan := make(chan common.Location)
						errChan := make(chan error)
						mockProtocol.On(`GetLocations`).Return([]common.Location{mockLocation}, nil).Once()
						protocolSubscription = common.NewSubscription(mockProtocol)
						mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
						mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
						mockLocation.Group.On(`GetLabel`).Return(locationLabel, nil).Once()
						go func() {
							loc, err := client.GetLocationByLabel(locationUnknownLabel)
							errChan <- err
							locChan <- loc
						}()
						unknownLocation := new(mocks.Location)
						unknownLocation.Group.On(`ID`).Return(locationUnknownID).Once()
						unknownLocation.Group.On(`GetLabel`).Return(locationUnknownLabel, nil).Once()
						_ = protocolSubscription.Write(common.EventNewLocation{Location: unknownLocation})
						Expect(<-errChan).NotTo(HaveOccurred())
						Expect(<-locChan).To(Equal(unknownLocation))
						close(done)
					})

				})

				Context("with zero timeout", func() {
					BeforeEach(func() {
						mockProtocol.On(`SetTimeout`, mock.AnythingOfType("*time.Duration")).Return().Once()
						client.SetTimeout(0)
						protocolSubscription = common.NewSubscription(mockProtocol)
						mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
						mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
					})

					It("should not timeout searching by ID", func(done Done) {
						time.AfterFunc(10*time.Millisecond, func() {
							close(done)
						})

						mockProtocol.On(`GetLocation`, locationUnknownID).Return(&mocks.Location{}, common.ErrNotFound).Once()
						_, err := client.GetLocationByID(locationUnknownID)
						Expect(err).NotTo(HaveOccurred())
					})

					It("should not timeout searching by label with results", func(done Done) {
						time.AfterFunc(100*time.Millisecond, func() {
							close(done)
						})

						mockProtocol.On(`GetLocations`).Return([]common.Location{mockLocation}, nil).Once()
						mockLocation.Group.On(`GetLabel`).Return(locationLabel, nil).Once()
						_, err := client.GetLocationByLabel(locationUnknownLabel)
						Expect(err).NotTo(HaveOccurred())
					})

					It("should not timeout searching by label without results", func(done Done) {
						time.AfterFunc(100*time.Millisecond, func() {
							close(done)
						})

						mockProtocol.On(`GetLocations`).Return(nil, common.ErrNotFound).Once()
						_, err := client.GetLocationByLabel(locationUnknownLabel)
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})
		})

		Context("with groups", func() {

			Context("finding a group", func() {
				It("should find it by ID", func() {
					mockProtocol.On(`GetGroup`, groupID).Return(mockGroup, nil).Once()
					loc, err := client.GetGroupByID(groupID)
					Expect(loc).To(Equal(mockGroup))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return an error when the ID is not known", func() {
					mockProtocol.On(`GetGroup`, groupUnknownID).Return(&mocks.Group{}, common.ErrNotFound).Once()
					mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
					mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
					_, err := client.GetGroupByID(groupUnknownID)
					Expect(err).To(MatchError(common.ErrNotFound))
				})

				It("should find it by label", func() {
					mockProtocol.On(`GetGroups`).Return([]common.Group{mockGroup}, nil).Once()
					mockGroup.On(`GetLabel`).Return(groupLabel, nil).Once()
					loc, err := client.GetGroupByLabel(groupLabel)
					Expect(loc).To(Equal(mockGroup))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return an error when the label is not known", func() {
					mockProtocol.On(`GetGroups`).Return([]common.Group{mockGroup}, nil).Once()
					mockGroup.On(`GetLabel`).Return(groupLabel, nil).Once()
					mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
					mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
					_, err := client.GetGroupByLabel(groupUnknownLabel)
					Expect(err).To(MatchError(common.ErrNotFound))
				})

				Context("when the group is added while searching", func() {

					It("should find it by ID", func(done Done) {
						grpChan := make(chan common.Group)
						errChan := make(chan error)
						mockProtocol.On(`GetGroup`, groupUnknownID).Return(&mocks.Group{}, common.ErrNotFound).Once()
						protocolSubscription = common.NewSubscription(mockProtocol)
						mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
						mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
						go func() {
							loc, err := client.GetGroupByID(groupUnknownID)
							errChan <- err
							grpChan <- loc
						}()
						unknownGroup := new(mocks.Group)
						unknownGroup.On(`ID`).Return(groupUnknownID).Once()
						_ = protocolSubscription.Write(common.EventNewGroup{Group: unknownGroup})
						Expect(<-errChan).NotTo(HaveOccurred())
						Expect(<-grpChan).To(Equal(unknownGroup))
						close(done)
					})

					It("should find it by label", func(done Done) {
						grpChan := make(chan common.Group)
						errChan := make(chan error)
						mockProtocol.On(`GetGroups`).Return([]common.Group{mockGroup}, nil).Once()
						protocolSubscription = common.NewSubscription(mockProtocol)
						mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
						mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
						mockGroup.On(`GetLabel`).Return(groupLabel, nil).Once()
						go func() {
							loc, err := client.GetGroupByLabel(groupUnknownLabel)
							errChan <- err
							grpChan <- loc
						}()
						unknownGroup := new(mocks.Group)
						unknownGroup.On(`ID`).Return(groupUnknownID).Once()
						unknownGroup.On(`GetLabel`).Return(groupUnknownLabel, nil).Once()
						_ = protocolSubscription.Write(common.EventNewGroup{Group: unknownGroup})
						Expect(<-errChan).NotTo(HaveOccurred())
						Expect(<-grpChan).To(Equal(unknownGroup))
						close(done)
					})

				})

				Context("with zero timeout", func() {
					BeforeEach(func() {
						mockProtocol.On(`SetTimeout`, mock.AnythingOfType("*time.Duration")).Return().Once()
						client.SetTimeout(0)
						protocolSubscription = common.NewSubscription(mockProtocol)
						mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
						mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
					})

					It("should not timeout searching by ID", func(done Done) {
						time.AfterFunc(10*time.Millisecond, func() {
							close(done)
						})

						mockProtocol.On(`GetGroup`, groupUnknownID).Return(&mocks.Group{}, common.ErrNotFound).Once()
						_, err := client.GetGroupByID(groupUnknownID)
						Expect(err).NotTo(HaveOccurred())
					})

					It("should not timeout searching by label with results", func(done Done) {
						time.AfterFunc(100*time.Millisecond, func() {
							close(done)
						})

						mockProtocol.On(`GetGroups`).Return([]common.Group{mockGroup}, nil).Once()
						mockGroup.On(`GetLabel`).Return(groupLabel, nil).Once()
						_, err := client.GetGroupByLabel(groupUnknownLabel)
						Expect(err).NotTo(HaveOccurred())
					})

					It("should not timeout searching by label without results", func(done Done) {
						time.AfterFunc(100*time.Millisecond, func() {
							close(done)
						})

						mockProtocol.On(`GetGroups`).Return(nil, common.ErrNotFound).Once()
						_, err := client.GetGroupByLabel(groupUnknownLabel)
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})
		})

		Context("with devices", func() {

			Context("finding a device", func() {
				It("should find it by ID", func() {
					mockProtocol.On(`GetDevice`, deviceID).Return(mockDevice, nil).Once()
					dev, err := client.GetDeviceByID(deviceID)
					Expect(dev).To(Equal(mockDevice))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return an error when the ID is not known", func() {
					mockProtocol.On(`GetDevice`, deviceUnknownID).Return(&mocks.Device{}, common.ErrNotFound).Once()
					protocolSubscription = common.NewSubscription(mockProtocol)
					mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
					mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
					_, err := client.GetDeviceByID(deviceUnknownID)
					Expect(err).To(MatchError(common.ErrNotFound))
				})

				It("should find it by label", func() {
					mockProtocol.On(`GetDevices`).Return([]common.Device{mockDevice}, nil).Once()
					mockDevice.On(`GetLabel`).Return(deviceLabel, nil).Once()
					dev, err := client.GetDeviceByLabel(deviceLabel)
					Expect(dev).To(Equal(mockDevice))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return an error when the label is not known", func() {
					mockProtocol.On(`GetDevices`).Return([]common.Device{mockDevice}, common.ErrNotFound).Once()
					protocolSubscription = common.NewSubscription(mockProtocol)
					mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
					mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
					mockDevice.On(`GetLabel`).Return(deviceLabel, nil).Once()
					_, err := client.GetDeviceByLabel(deviceUnknownLabel)
					Expect(err).To(MatchError(common.ErrNotFound))
				})

				Context("when the device is added while searching", func() {

					It("should find it by ID", func(done Done) {
						devChan := make(chan common.Device)
						errChan := make(chan error)
						mockProtocol.On(`GetDevice`, deviceUnknownID).Return(&mocks.Device{}, common.ErrNotFound).Once()
						protocolSubscription = common.NewSubscription(mockProtocol)
						mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
						mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
						go func() {
							loc, err := client.GetDeviceByID(deviceUnknownID)
							errChan <- err
							devChan <- loc
						}()
						unknownDevice := new(mocks.Device)
						unknownDevice.On(`ID`).Return(deviceUnknownID).Once()
						_ = protocolSubscription.Write(common.EventNewDevice{Device: unknownDevice})
						Expect(<-errChan).NotTo(HaveOccurred())
						Expect(<-devChan).To(Equal(unknownDevice))
						close(done)
					})

					It("should find it by label", func(done Done) {
						devChan := make(chan common.Device)
						errChan := make(chan error)
						mockProtocol.On(`GetDevices`).Return([]common.Device{mockDevice}, nil).Once()
						protocolSubscription = common.NewSubscription(mockProtocol)
						mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
						mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
						mockDevice.On(`GetLabel`).Return(deviceLabel, nil).Once()
						go func() {
							loc, err := client.GetDeviceByLabel(deviceUnknownLabel)
							errChan <- err
							devChan <- loc
						}()
						unknownDevice := new(mocks.Device)
						unknownDevice.On(`ID`).Return(deviceUnknownID).Once()
						unknownDevice.On(`GetLabel`).Return(deviceUnknownLabel, nil).Once()
						_ = protocolSubscription.Write(common.EventNewDevice{Device: unknownDevice})
						Expect(<-errChan).NotTo(HaveOccurred())
						Expect(<-devChan).To(Equal(unknownDevice))
						close(done)
					})

				})

				Context("with zero timeout", func() {
					BeforeEach(func() {
						mockProtocol.On(`SetTimeout`, mock.AnythingOfType("*time.Duration")).Return().Once()
						client.SetTimeout(0)
						protocolSubscription = common.NewSubscription(mockProtocol)
						mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
						mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
					})

					It("should not timeout searching by ID", func(done Done) {
						time.AfterFunc(10*time.Millisecond, func() {
							close(done)
						})

						mockProtocol.On(`GetDevice`, deviceUnknownID).Return(&mocks.Device{}, common.ErrNotFound).Once()
						_, err := client.GetDeviceByID(deviceUnknownID)
						Expect(err).NotTo(HaveOccurred())
					})

					It("should not timeout searching by label with results", func(done Done) {
						time.AfterFunc(10*time.Millisecond, func() {
							close(done)
						})

						mockProtocol.On(`GetDevices`).Return([]common.Device{mockDevice}, nil).Once()
						mockDevice.On(`GetLabel`).Return(deviceLabel, nil).Once()
						_, err := client.GetDeviceByLabel(deviceUnknownLabel)
						Expect(err).NotTo(HaveOccurred())
					})

					It("should not timeout searching by label without results", func(done Done) {
						time.AfterFunc(10*time.Millisecond, func() {
							close(done)
						})

						mockProtocol.On(`GetDevices`).Return(nil, common.ErrNotFound).Once()
						_, err := client.GetDeviceByLabel(deviceUnknownLabel)
						Expect(err).NotTo(HaveOccurred())
					})

				})
			})

			It("should not return any lights", func() {
				mockProtocol.On(`GetDevices`).Return([]common.Device{mockDevice}, nil).Once()
				lights, err := client.GetLights()
				Expect(len(lights)).To(Equal(0))
				Expect(err).To(MatchError(common.ErrNotFound))
			})

			Context("with lights", func() {

				It("should return only lights", func() {
					mockProtocol.On(`GetDevices`).Return([]common.Device{mockDevice, mockLight}, nil).Once()
					lights, err := client.GetLights()
					Expect(len(lights)).To(Equal(1))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return it by ID when known", func() {
					mockProtocol.On(`GetDevice`, lightID).Return(mockLight, nil).Once()
					light, err := client.GetLightByID(lightID)
					Expect(light).To(Equal(mockLight))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should not return a known device by ID if it is not a light", func() {
					mockProtocol.On(`GetDevice`, deviceID).Return(mockDevice, nil).Once()
					light, err := client.GetLightByID(deviceID)
					Expect(light).To(BeNil())
					Expect(err).To(HaveOccurred())
				})

				It("should return it by label when known", func() {
					mockProtocol.On(`GetDevices`).Return([]common.Device{mockDevice, mockLight}, nil).Once()
					mockDevice.On(`GetLabel`).Return(deviceLabel, nil).Once()
					mockLight.Device.On(`GetLabel`).Return(lightLabel, nil).Once()
					light, err := client.GetLightByLabel(lightLabel)
					Expect(light).To(Equal(mockLight))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should not return a known device by label if it is not a light", func() {
					mockProtocol.On(`GetDevices`).Return([]common.Device{mockDevice, mockLight}, nil).Once()
					mockDevice.On(`GetLabel`).Return(deviceLabel, nil).Once()
					mockLight.Device.On(`GetLabel`).Return(lightLabel, nil).Once()
					light, err := client.GetLightByLabel(deviceLabel)
					Expect(light).To(BeNil())
					Expect(err).To(MatchError(common.ErrDeviceInvalidType))
				})

				Context("with zero timeout", func() {
					BeforeEach(func() {
						mockProtocol.On(`SetTimeout`, mock.AnythingOfType("*time.Duration")).Return().Once()
						client.SetTimeout(0)
						protocolSubscription = common.NewSubscription(mockProtocol)
						mockProtocol.SubscriptionTarget.On(`NewSubscription`).Return(protocolSubscription, nil).Once()
						mockProtocol.SubscriptionTarget.On(`CloseSubscription`, protocolSubscription).Return(nil).Once()
					})

					It("should not timeout searching by ID", func(done Done) {
						time.AfterFunc(10*time.Millisecond, func() {
							close(done)
						})

						mockProtocol.On(`GetDevice`, deviceUnknownID).Return(&mocks.Device{}, common.ErrNotFound).Once()
						_, err := client.GetLightByID(deviceUnknownID)
						Expect(err).NotTo(HaveOccurred())
					})

					It("should not timeout searching by label with results", func(done Done) {
						time.AfterFunc(10*time.Millisecond, func() {
							close(done)
						})

						mockProtocol.On(`GetDevices`).Return([]common.Device{mockDevice, mockLight}, nil).Once()
						mockDevice.On(`GetLabel`).Return(deviceLabel, nil).Once()
						mockLight.Device.On(`GetLabel`).Return(lightLabel, nil).Once()
						_, err := client.GetLightByLabel(deviceUnknownLabel)
						Expect(err).NotTo(HaveOccurred())
					})

					It("should not timeout searching by label without results", func(done Done) {
						time.AfterFunc(10*time.Millisecond, func() {
							close(done)
						})

						mockProtocol.On(`GetDevices`).Return(nil, common.ErrNotFound).Once()
						_, err := client.GetLightByLabel(deviceUnknownLabel)
						Expect(err).NotTo(HaveOccurred())
					})
				})

			})

		})

	})

})
