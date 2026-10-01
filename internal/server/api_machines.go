package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tanq16/anbu/internal/awscred"
	"github.com/tanq16/anbu/internal/awsx"
	"github.com/tanq16/anbu/internal/inventory"
	"github.com/tanq16/anbu/internal/machine"
	"github.com/tanq16/anbu/internal/pricing"
	"github.com/tanq16/anbu/internal/scaffold"
	"github.com/tanq16/anbu/internal/sizing"
	"github.com/tanq16/anbu/internal/vault"
)

const optionsTTL = time.Hour

var machineName = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

type optionsEntry struct {
	mu       sync.Mutex
	prices   *pricing.Table
	resolver *sizing.Resolver
	expires  time.Time
}

type optionsCache struct {
	mu      sync.Mutex
	entries map[string]*optionsEntry
}

func newOptionsCache() *optionsCache {
	return &optionsCache{entries: map[string]*optionsEntry{}}
}

func (s *Server) priced(ctx context.Context, c *awsx.Clients) (*pricing.Table, *sizing.Resolver, error) {
	key := c.Account + "/" + c.Region
	s.options.mu.Lock()
	e, ok := s.options.entries[key]
	if !ok {
		e = &optionsEntry{}
		s.options.entries[key] = e
	}
	s.options.mu.Unlock()

	e.mu.Lock()
	defer e.mu.Unlock()
	if e.prices != nil && time.Now().Before(e.expires) {
		return e.prices, e.resolver, nil
	}
	prices, err := pricing.Fetch(ctx, c)
	if err != nil {
		return nil, nil, err
	}
	resolver, err := sizing.NewResolver(ctx, c, prices)
	if err != nil {
		return nil, nil, err
	}
	e.prices, e.resolver, e.expires = prices, resolver, time.Now().Add(optionsTTL)
	return prices, resolver, nil
}

type resolvedOption struct {
	Type    string             `json:"type"`
	Hourly  float64            `json:"hourly"`
	Monthly map[string]float64 `json:"monthly"`
}

type machineOptions struct {
	Account    string                                                `json:"account"`
	Region     string                                                `json:"region"`
	Shapes     []string                                              `json:"shapes"`
	Disks      []int                                                 `json:"disks"`
	Arches     []string                                              `json:"arches"`
	Classes    []sizing.Class                                        `json:"classes"`
	GP3GBMonth float64                                               `json:"gp3_gb_month"`
	IPv4Hourly float64                                               `json:"ipv4_hourly"`
	Types      map[string]map[string]map[sizing.Class]resolvedOption `json:"types"`
	Stopped    map[string]float64                                    `json:"stopped"`
}

type machineRequest struct {
	Source  awscred.Source `json:"source"`
	Name    string         `json:"name"`
	Arch    string         `json:"arch"`
	Shape   string         `json:"shape"`
	Class   sizing.Class   `json:"class"`
	Disk    int            `json:"disk"`
	Confirm string         `json:"confirm"`
}

func (s *Server) routeMachines() {
	s.mux.HandleFunc("GET /api/aws/machines", s.handleListMachines)
	s.mux.HandleFunc("GET /api/aws/machines/options", s.handleMachineOptions)
	s.mux.HandleFunc("GET /api/aws/machines/{name}/status", s.handleMachineStatus)
	s.mux.HandleFunc("POST /api/aws/machines", s.handleCreateMachine)
	s.mux.HandleFunc("POST /api/aws/machines/{name}/start", s.handleMachineLifecycle("start"))
	s.mux.HandleFunc("POST /api/aws/machines/{name}/stop", s.handleMachineLifecycle("stop"))
	s.mux.HandleFunc("POST /api/aws/machines/{name}/modify", s.handleModifyMachine)
	s.mux.HandleFunc("DELETE /api/aws/machines/{name}", s.handleMachineLifecycle("remove"))
	s.mux.HandleFunc("GET /api/aws/scaffold", s.handleScaffoldStatus)
	s.mux.HandleFunc("POST /api/aws/scaffold", s.handleScaffoldJob("setup"))
	s.mux.HandleFunc("DELETE /api/aws/scaffold", s.handleScaffoldJob("teardown"))
	s.mux.HandleFunc("PUT /api/aws/scaffold/key", s.handleAdoptScaffoldKey)
}

func (s *Server) queryClients(w http.ResponseWriter, r *http.Request) (*awsx.Clients, bool) {
	src := querySource(r)
	if err := s.creds.Validate(src); err != nil {
		s.fail(w, r, err)
		return nil, false
	}
	c, err := s.clients(r.Context(), src)
	if err != nil {
		s.failRemote(w, r, err)
		return nil, false
	}
	return c, true
}

func (s *Server) handleListMachines(w http.ResponseWriter, r *http.Request) {
	c, ok := s.queryClients(w, r)
	if !ok {
		return
	}
	prices, _, err := s.priced(r.Context(), c)
	if err != nil {
		log.Warn().Err(err).Str("account", c.Account).Str("region", c.Region).Msg("listing machines without prices")
		prices = nil
	}
	machines, err := inventory.FromAPI(r.Context(), c, prices)
	if err != nil {
		s.failRemote(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, machines)
}

func (s *Server) handleMachineOptions(w http.ResponseWriter, r *http.Request) {
	c, ok := s.queryClients(w, r)
	if !ok {
		return
	}
	prices, resolver, err := s.priced(r.Context(), c)
	if err != nil {
		s.failRemote(w, r, err)
		return
	}
	opts := machineOptions{
		Account:    c.Account,
		Region:     c.Region,
		Disks:      sizing.DiskSizes(),
		Arches:     sizing.Arches(),
		Classes:    sizing.Classes(),
		GP3GBMonth: prices.GP3GBMonth(),
		IPv4Hourly: prices.PublicIPv4Hourly(),
		Types:      map[string]map[string]map[sizing.Class]resolvedOption{},
		Stopped:    map[string]float64{},
	}
	for _, disk := range opts.Disks {
		opts.Stopped[strconv.Itoa(disk)] = prices.StoppedMonth(disk)
	}
	for _, shape := range sizing.Shapes() {
		opts.Shapes = append(opts.Shapes, shape.ID())
	}
	for _, arch := range opts.Arches {
		byShape := map[string]map[sizing.Class]resolvedOption{}
		for _, shape := range sizing.Shapes() {
			byClass := map[sizing.Class]resolvedOption{}
			for _, class := range opts.Classes {
				res, err := resolver.Resolve(shape, arch, class)
				if err != nil {
					s.fail(w, r, err)
					return
				}
				if res == nil {
					continue
				}
				opt := resolvedOption{Type: res.InstanceType, Hourly: res.HourlyUSD, Monthly: map[string]float64{}}
				for _, disk := range opts.Disks {
					opt.Monthly[strconv.Itoa(disk)] = prices.RunningMonth(res.HourlyUSD, disk)
				}
				byClass[class] = opt
			}
			if len(byClass) > 0 {
				byShape[shape.ID()] = byClass
			}
		}
		opts.Types[arch] = byShape
	}
	writeJSON(w, http.StatusOK, opts)
}

func (s *Server) handleMachineStatus(w http.ResponseWriter, r *http.Request) {
	c, ok := s.queryClients(w, r)
	if !ok {
		return
	}
	report, err := machine.Status(r.Context(), c, s.vault, s.knownHosts, r.PathValue("name"))
	if err != nil {
		s.failRemote(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"report": report})
}

func (s *Server) machineInput(w http.ResponseWriter, r *http.Request, bodyRequired bool) (machineRequest, error) {
	var in machineRequest
	if bodyRequired || r.ContentLength != 0 {
		if err := readJSON(w, r, &in); err != nil {
			return machineRequest{}, err
		}
	}
	if in.Source.Profile == "" && in.Source.Inline == nil {
		in.Source = querySource(r)
	}
	if err := s.creds.Validate(in.Source); err != nil {
		return machineRequest{}, err
	}
	return in, nil
}

func validName(name string) error {
	if !machineName.MatchString(name) {
		return badRequest("machine name must be 1 to 64 letters, digits, '-', '_', or '.'")
	}
	return nil
}

func validShapeClass(shapeID string, class sizing.Class) (sizing.Shape, error) {
	shape, ok := sizing.ParseShape(shapeID)
	if !ok {
		return sizing.Shape{}, badRequest("shape %q is not one of the offered shapes", shapeID)
	}
	if !slices.Contains(sizing.Classes(), class) {
		return sizing.Shape{}, badRequest("class must be burstable or dedicated")
	}
	return shape, nil
}

func (s *Server) enqueue(w http.ResponseWriter, kind, label string, run func(ctx context.Context) (any, error)) {
	job := s.jobs.Enqueue(kind, label, func(ctx context.Context) (any, error) {
		result, err := run(ctx)
		if errors.Is(err, awscred.ErrLoginRequired) {
			err = awscred.ErrLoginRequired
		}
		return result, err
	})
	log.Info().Str("job", job.ID).Str("kind", kind).Str("label", label).Msg("job queued")
	writeJSON(w, http.StatusAccepted, job)
}

func (s *Server) handleCreateMachine(w http.ResponseWriter, r *http.Request) {
	in, err := s.machineInput(w, r, true)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if err := validName(in.Name); err != nil {
		s.fail(w, r, err)
		return
	}
	if !slices.Contains(sizing.Arches(), in.Arch) {
		s.fail(w, r, badRequest("arch must be arm64 or x86_64"))
		return
	}
	shape, err := validShapeClass(in.Shape, in.Class)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if !slices.Contains(sizing.DiskSizes(), in.Disk) {
		s.fail(w, r, badRequest("disk must be one of %v GB", sizing.DiskSizes()))
		return
	}
	src, timezone := in.Source, s.settings.Get().MachineTimezone
	label := fmt.Sprintf("create %s on %s", in.Name, src.Label())
	s.enqueue(w, "machine.create", label, func(ctx context.Context) (any, error) {
		c, err := s.clients(ctx, src)
		if err != nil {
			return nil, err
		}
		_, resolver, err := s.priced(ctx, c)
		if err != nil {
			return nil, err
		}
		target, err := resolver.Resolve(shape, in.Arch, in.Class)
		if err != nil {
			return nil, err
		}
		if target == nil {
			return nil, fmt.Errorf("no %s instance type offers %s on %s in %s", in.Class, in.Shape, in.Arch, c.Region)
		}
		return machine.Create(ctx, c, s.vault, machine.CreateConfig{
			Name:         in.Name,
			InstanceType: target.InstanceType,
			Arch:         target.Arch,
			VCPU:         target.VCPU,
			MemoryGB:     target.MemoryGB,
			DiskGB:       in.Disk,
			Timezone:     timezone,
		})
	})
}

func (s *Server) handleModifyMachine(w http.ResponseWriter, r *http.Request) {
	in, err := s.machineInput(w, r, true)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	name := r.PathValue("name")
	if err := validName(name); err != nil {
		s.fail(w, r, err)
		return
	}
	shape, err := validShapeClass(in.Shape, in.Class)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	src := in.Source
	label := fmt.Sprintf("modify %s to %s %s on %s", name, in.Shape, in.Class, src.Label())
	s.enqueue(w, "machine.modify", label, func(ctx context.Context) (any, error) {
		c, err := s.clients(ctx, src)
		if err != nil {
			return nil, err
		}
		current, err := c.FindInstance(ctx, name)
		if err != nil {
			return nil, err
		}
		if current == nil {
			return nil, fmt.Errorf("%w: no machine named %q in %s", machine.ErrNotFound, name, c.Region)
		}
		_, resolver, err := s.priced(ctx, c)
		if err != nil {
			return nil, err
		}
		target, err := resolver.Resolve(shape, current.Arch, in.Class)
		if err != nil {
			return nil, err
		}
		if target == nil {
			return nil, fmt.Errorf("no %s instance type offers %s on %s in %s", in.Class, in.Shape, current.Arch, c.Region)
		}
		vcpu, memoryGB, arches, err := resolver.Details(ctx, target.InstanceType)
		if err != nil {
			return nil, err
		}
		return machine.Modify(ctx, c, machine.ModifyConfig{
			Name:         name,
			InstanceType: target.InstanceType,
			VCPU:         vcpu,
			MemoryGB:     memoryGB,
			Arches:       arches,
		})
	})
}

func (s *Server) handleMachineLifecycle(action string) http.HandlerFunc {
	ops := map[string]func(context.Context, *awsx.Clients, string) (*machine.Info, error){
		"start":  machine.Start,
		"stop":   machine.Stop,
		"remove": machine.Remove,
	}
	op := ops[action]
	return func(w http.ResponseWriter, r *http.Request) {
		in, err := s.machineInput(w, r, action == "remove")
		if err != nil {
			s.fail(w, r, err)
			return
		}
		name := r.PathValue("name")
		if err := validName(name); err != nil {
			s.fail(w, r, err)
			return
		}
		if action == "remove" && in.Confirm != name {
			s.fail(w, r, badRequest("confirm must equal the machine name %q", name))
			return
		}
		src := in.Source
		s.enqueue(w, "machine."+action, fmt.Sprintf("%s %s on %s", action, name, src.Label()), func(ctx context.Context) (any, error) {
			c, err := s.clients(ctx, src)
			if err != nil {
				return nil, err
			}
			info, err := op(ctx, c, name)
			if err == nil && action == "remove" {
				if _, err := s.knownHosts.Remove(c.HostKeyAlias(name)); err != nil {
					log.Warn().Err(err).Str("machine", name).Msg("failed to forget host key")
				}
			}
			return info, err
		})
	}
}

type adoptKeyRequest struct {
	Source     awscred.Source `json:"source"`
	PrivateKey string         `json:"private_key"`
}

func (s *Server) handleAdoptScaffoldKey(w http.ResponseWriter, r *http.Request) {
	var in adoptKeyRequest
	if err := readJSON(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	if in.Source.Profile == "" && in.Source.Inline == nil {
		in.Source = querySource(r)
	}
	if err := s.creds.Validate(in.Source); err != nil {
		s.fail(w, r, err)
		return
	}
	if strings.TrimSpace(in.PrivateKey) == "" {
		s.fail(w, r, badRequest("private_key is required"))
		return
	}
	c, err := s.clients(r.Context(), in.Source)
	if err != nil {
		s.failRemote(w, r, err)
		return
	}
	status, err := scaffold.AdoptKey(r.Context(), c, s.vault, in.PrivateKey)
	if err != nil {
		s.failRemote(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleScaffoldStatus(w http.ResponseWriter, r *http.Request) {
	c, ok := s.queryClients(w, r)
	if !ok {
		return
	}
	status, err := scaffold.Inspect(r.Context(), c, s.vault)
	if err != nil {
		s.failRemote(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleScaffoldJob(action string) http.HandlerFunc {
	ops := map[string]func(context.Context, *awsx.Clients, *vault.Store) ([]scaffold.Event, error){
		"setup":    scaffold.Setup,
		"teardown": scaffold.Teardown,
	}
	op := ops[action]
	return func(w http.ResponseWriter, r *http.Request) {
		in, err := s.machineInput(w, r, false)
		if err != nil {
			s.fail(w, r, err)
			return
		}
		src := in.Source
		s.enqueue(w, "scaffold."+action, fmt.Sprintf("%s scaffold on %s", action, src.Label()), func(ctx context.Context) (any, error) {
			c, err := s.clients(ctx, src)
			if err != nil {
				return nil, err
			}
			events, err := op(ctx, c, s.vault)
			if len(events) == 0 {
				return nil, err
			}
			return events, err
		})
	}
}
