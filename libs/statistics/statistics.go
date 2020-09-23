package statistics

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/eatools/gservice/libs/logmod"
	"github.com/sirupsen/logrus"

	client "github.com/influxdata/influxdb1-client"
	"github.com/spf13/cast"
)

var (
	defaultStat *Statistics
	defaultOpt  = &StatisticsOpt{
		Delta:     -1 * time.Minute, // 默认取前一分钟的数据，(-1*time.Minute)
		Tags:      nil,              // 默认添加的tags，在使用CountItem时，不能出现默认的tags中的key
		DataCount: 5,                //
		NodeCount: 10000,            // chan的长度，默认是10000
	}
	errFileds = errors.New("invalid_fields")
	errTags   = errors.New("invalid_tags")
)

// NewStatisticsWithDefault default
func NewStatisticsWithDefault(url, db string) error {

	opt := defaultOpt
	defaultStat = NewStatistics(opt)
	log.Printf("influxdb..., %v, %v, %v\n", time.Minute, url, db)
	err := InfluxDB(defaultStat, time.Minute, url, db, "", "")
	//err := InfluxDB(defaultStat, 10*time.Second, url, db, "", "")
	if err != nil {
		return err
	}
	return nil
}

// CountOne service need this
func CountOne(name, tp string, tags map[string]string, ts ...time.Time) error {
	return defaultStat.CountOne(name, tp, tags, ts...)
}

func CountOneField(name, tp string, tags map[string]string, fields map[string]interface{}) error {
	return defaultStat.CountOneField(name, tp, tags, fields)
}

// Registry registry
type Registry interface {
	Echo(f func(n Noder))
}

// Noder node
type Noder interface {
	Typ() string
	Name() string
	Tags() map[string]string
	Fields() map[string]interface{}
	Time() time.Time
}

type reporter struct {
	registry Registry
	interval time.Duration

	address  string
	database string
	username string
	password string

	client *client.Client
	delta  int
}

// InfluxDB starts a InfluxDB reporter which will post the metrics from the given registry at each d interval.
func InfluxDB(r Registry, d time.Duration, address, database, username, password string) error {
	return InfluxDBWithTags(r, d, address, database, username, password)
}

// InfluxDBWithTags starts a InfluxDB reporter which will post the metrics from the given registry at each d interval with the specified tags
func InfluxDBWithTags(r Registry, d time.Duration, address, database, username, password string) error {
	rep := &reporter{
		registry: r,
		interval: d,
		address:  address,
		database: database,
		username: username,
		password: password,
		delta:    10000,
	}
	if err := rep.makeClient(); err != nil {
		log.Printf("unable to make InfluxDB client. err=%v", err)
		return err
	}

	go rep.run()
	return nil
}

func (r *reporter) makeClient() error {
	host, err := url.Parse(r.address)
	if err != nil {
		log.Fatal(err)
	}
	r.client, err = client.NewClient(client.Config{
		URL:      *host,
		Username: r.username,
		Password: r.password,
	})

	return err
}

func (r *reporter) run() {
	intervalTicker := time.Tick(r.interval)
	pingTicker := time.Tick(time.Second * 500)

	for {
		select {
		case <-intervalTicker:
			if err := r.send(); err != nil {
				log.Printf("unable to send metrics to InfluxDB. err=%v", err)
			}
		case <-pingTicker:
			_, _, err := r.client.Ping()
			if err != nil {
				log.Printf("got error while sending a ping to InfluxDB, trying to recreate client. err=%v", err)

				if err = r.makeClient(); err != nil {
					log.Printf("unable to make InfluxDB client. err=%v", err)
				}
			}
		}
	}
}

func (r *reporter) send() error {
	var sspPoints []client.Point
	var dspPoints []client.Point

	fmt.Println("------ start send ------")

	r.registry.Echo(func(n Noder) {
		pt := client.Point{
			Measurement: n.Name(),
			Tags:        n.Tags(),
			Time:        n.Time(),
			Fields:      n.Fields(),
		}

		switch n.Name() {
		case "ssp":
			sspPoints = append(sspPoints, pt)
		case "dsp":
			dspPoints = append(dspPoints, pt)
		default:
		}

		if len(sspPoints) >= r.delta {
			sspbps := client.BatchPoints{
				Points:   sspPoints,
				Database: r.database,
				//RetentionPolicy: "default",
			}

			fmt.Println("max more than r.delta : ", len(sspPoints))
			_, err := r.client.Write(sspbps)
			if err != nil {
				logmod.Type("").WithFields(logrus.Fields{"wirteDspBsp len > 10000": err.Error()}).Error("write dspbps err")
			}
			sspPoints = sspPoints[:0]
		}

		if len(dspPoints) >= r.delta {
			dspbps := client.BatchPoints{
				Points:   dspPoints,
				Database: r.database,
			}
			fmt.Println("dspPoint max more than r.delta : ", len(dspPoints))
			_, err := r.client.Write(dspbps)
			if err != nil {
				logmod.Type("").WithFields(logrus.Fields{"wirteDspBsp len > 10000": err.Error()}).Error("write dspbps err")
			}
			dspPoints = dspPoints[:0]
		}
	})

	if len(sspPoints) > 0 {
		fmt.Println("sspPoints len : ", len(sspPoints))
		sspbps := client.BatchPoints{
			Points:   sspPoints,
			Database: r.database,
			//RetentionPolicy: "default",
		}
		_, err := r.client.Write(sspbps)
		if err != nil {
			//panic(err)
			logmod.Type("").WithFields(logrus.Fields{"wirteSspBsp": err.Error()}).Error("write Sspbps err")
			return err
		}
	}

	if len(dspPoints) > 0 {
		fmt.Println("dspPoints len : ", len(dspPoints))
		dspbps := client.BatchPoints{
			Points:   dspPoints,
			Database: r.database,
		}
		_, err := r.client.Write(dspbps)
		if err != nil {
			//panic(err)
			logmod.Type("").WithFields(logrus.Fields{"wirteDspBsp len > 0": err.Error()}).Error("write dspbps err")
			return err
		}
	}

	return nil
}

var (
	value = "_value_"
)

// Statistics 统计
type Statistics struct {
	delta    time.Duration // 默认取前一分钟的数据，(-1*time.Minute)
	tags     map[string]string
	NodeData map[int64]map[string]*Node
	NodeCh   chan *Node
	pool     *NodePool
	mu       sync.Mutex
}

// StatisticsOpt 选项
type StatisticsOpt struct {
	Delta     time.Duration     // 默认取前一分钟的数据，(-1*time.Minute)
	Tags      map[string]string // 每条记录都会默认添加的tags，在使用CountItem时，不能出现默认的tags中的key
	DataCount int               // 保留多久的数据，默认是2
	NodeCount int               // chan的长度，默认是10000

}

// NodeList x
type NodeList struct {
	timeStamp int64
	nodes     []*Node
}

// Node node
type Node struct {
	typ    string
	name   string
	tags   map[string]string
	fields map[string]interface{}
	itime  time.Time
	// Value  int64
}

// Type type
func (n *Node) Typ() string {
	return n.typ
}

// Name name
func (n *Node) Name() string {
	return n.name
}

// Tags tags
func (n *Node) Tags() map[string]string {
	return n.tags
}

// Fields fields
func (n *Node) Fields() map[string]interface{} {
	return n.fields
}

// Time time
func (n *Node) Time() time.Time {
	return n.itime
}

// NewStatistics 初始化
func NewStatistics(opt *StatisticsOpt) *Statistics {
	if opt == nil {
		opt = defaultOpt
	}

	if opt.Delta == 0 {
		opt.Delta = defaultOpt.Delta
	}

	if opt.DataCount < 2 || opt.DataCount > 10 {
		opt.DataCount = defaultOpt.DataCount
	}

	if opt.NodeCount < 10 || opt.NodeCount > 200000 {
		opt.NodeCount = defaultOpt.NodeCount
	}

	s := &Statistics{
		delta:    opt.Delta,
		tags:     opt.Tags,
		NodeData: make(map[int64]map[string]*Node),
		NodeCh:   make(chan *Node, opt.NodeCount),
		pool:     NewNodePool(),
	}

	if s.tags == nil {
		s.tags = make(map[string]string)
	}

	go s.collect()
	go s.clean()

	return s
}

// SetTags 设置tags
func (s *Statistics) SetTags(tags map[string]string) {
	s.tags = tags
}

func (s *Statistics) clean() {
	for range time.Tick(time.Minute * 5) {
		now := time.Now()
		t := getTimeStamp(now, 0)
		s.mu.Lock()
		for k := range s.NodeData {
			if t-k < 5*60 {
				continue
			}
			delete(s.NodeData, k)
		}
		s.mu.Unlock()
	}
}

// CountOne Add的一个封装，方便使用
func (s *Statistics) CountOne(name, tp string, tags map[string]string, ts ...time.Time) error {
	var t time.Time
	if len(ts) == 0 {
		t = time.Now()
	} else {
		t = ts[0]
	}
	return s.Add(name, tp, tags, nil, 1, t)
}

// Add 增加记录，参数最全
func (s *Statistics) Add(name, tp string, tags map[string]string, fields map[string]interface{}, val interface{}, t time.Time) error {
	node := s.pool.Get()
	node.name = name
	node.itime = t
	node.typ = tp

	for k, v := range tags {
		if _, ok := s.tags[k]; ok {
			s.pool.Put(node)
			return errTags
		}
		node.tags[k] = v
	}

	// 默认添加一个字段： value_
	if _, ok := node.fields[value]; ok {
		s.pool.Put(node)
		return errFileds
	}
	node.fields[value] = val
	for k, v := range fields {
		node.fields[k] = v
	}

	s.NodeCh <- node
	return nil
}

func (s *Statistics) collect() {
	// start := time.Now()

	for node := range s.NodeCh {
		t := getTimeStamp(node.itime, 0)
		point := node.Typ() + "|" + NewTags(node.tags).String()

		s.mu.Lock()
		if _, ok := s.NodeData[t]; !ok {
			s.NodeData[t] = make(map[string]*Node)
		}
		if _, ok := s.NodeData[t][point]; !ok {
			s.NodeData[t][point] = node
		} else {
			for k, v := range node.fields {
				switch v2 := v.(type) {
				case int8, int16, int32, int64, int:
					v1 := cast.ToInt(s.NodeData[t][point].fields[k])
					s.NodeData[t][point].fields[k] = v1 + cast.ToInt(v2)
				case float32, float64:
					v1 := cast.ToFloat64(s.NodeData[t][point].fields[k])
					s.NodeData[t][point].fields[k] = v1 + cast.ToFloat64(v2)
				default:
				}
			}
		}
		s.mu.Unlock()
	}
	// fmt.Printf("collect time:%v", time.Since(start).Seconds())
}

// Echo 遍历处理
func (s *Statistics) Echo(f func(n Noder)) {
	// get key
	now := time.Now()
	t := getTimeStamp(now, s.delta)

	m, ok := s.NodeData[t]
	if !ok {
		return
	}

	s.mu.Lock()
	delete(s.NodeData, t)
	s.mu.Unlock()

	//遍历node
	for _, n := range m {
		for k, v := range s.tags {
			n.tags[k] = v
		}
		f(n)
		s.pool.Put(n)
	}

}

func getTimeStamp(t time.Time, delta time.Duration) int64 {
	return t.Add(delta).Truncate(time.Minute).Unix()
}

// NodePool pool
type NodePool struct {
	pool sync.Pool
}

func NewNodePool() *NodePool {
	p := &NodePool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Node{
					tags:   make(map[string]string),
					fields: make(map[string]interface{}),
				}
			},
		},
	}

	return p
}

func (p *NodePool) Get() *Node {
	node := p.pool.Get().(*Node)

	if len(node.tags) != 0 {
		node.tags = make(map[string]string)
	}
	if len(node.fields) != 0 {
		node.fields = make(map[string]interface{})
	}

	return node
}

func (p *NodePool) Put(node *Node) {
	p.pool.Put(node)
}

// pool
var tagPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(nil)
	},
}

type Tags []Tag

type Tag struct {
	Key   []byte
	Value []byte
}

// NewTag returns a new Tag.
func NewTag(key, value []byte) Tag {
	return Tag{
		Key:   key,
		Value: value,
	}
}

// NewTags returns a new Tags from a map.
func NewTags(m map[string]string) Tags {
	if len(m) == 0 {
		return nil
	}
	a := make(Tags, 0, len(m))
	for k, v := range m {
		a = append(a, NewTag([]byte(k), []byte(v)))
	}
	sort.Sort(a)
	return a
}

func (a Tags) String() string {
	var buf = tagPool.Get().(*bytes.Buffer)
	buf.Reset()
	for i := range a {
		buf.Write(a[i].Key)
		buf.Write(a[i].Value)
		buf.WriteByte('|')
	}
	s := buf.String()
	tagPool.Put(buf)
	return s
}

func (a Tags) Len() int           { return len(a) }
func (a Tags) Less(i, j int) bool { return bytes.Compare(a[i].Key, a[j].Key) == -1 }
func (a Tags) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
