package dnshandler

import (
	"context"
	"github.com/jclab-joseph/dnsever-rfc2136-bridge/dnsever"
	"github.com/jclab-joseph/dnsever-rfc2136-bridge/model"
	"github.com/miekg/dns"
	"log"
	"strings"
	"time"
)

const (
	headerSize = 12

	// Header.Bits
	_QR = 1 << 15 // query/response (response=1)
	_AA = 1 << 10 // authoritative
	_TC = 1 << 9  // truncated
	_RD = 1 << 8  // recursion desired
	_RA = 1 << 7  // recursion available
	_Z  = 1 << 6  // Z
	_AD = 1 << 5  // authenticated data
	_CD = 1 << 4  // checking disabled
)

type Handler struct {
	domains  map[string]*domainZone
	tsigAuth map[string]*tsigAuth
}

func (h *Handler) ApplyAuthFile(authFile *model.AuthFile) {
	h.domains = make(map[string]*domainZone)
	h.tsigAuth = make(map[string]*tsigAuth)

	for _, domain := range authFile.Domains {
		z := &domainZone{
			Zone:     domain.Zone,
			Upstream: domain.Upstream,
			Tsig:     make(map[string]*tsigAuth),
		}
		if !strings.HasSuffix(z.Zone, ".") {
			z.Zone += "."
		}
		for _, auth := range domain.Tsig {
			item := &tsigAuth{
				TsigAuth: auth,
				Zone:     z,
			}
			h.tsigAuth[auth.Name] = item
			z.Tsig[auth.Name] = item
		}
		z.Client = dnsever.NewClient(domain.ClientId, domain.ClientSecret)
		h.domains[domain.Zone] = z
	}
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	tsig := r.IsTsig()
	zone, authOk := h.validateTsig(w, tsig, r.Opcode == dns.OpcodeQuery)
	if !authOk {
		m.SetRcode(r, dns.RcodeNotAuth)
		_ = w.WriteMsg(m)
		return
	}
	zone = h.validateZone(zone, r)
	if zone == nil {
		log.Printf("authenticate failed with wrong zone")
		m.SetRcode(r, dns.RcodeServerFailure)
		_ = w.WriteMsg(m)
		return
	}

	switch r.Opcode {
	case dns.OpcodeQuery:
		h.serveQuery(zone, w, r, m)

	case dns.OpcodeUpdate:
		h.serveUpdate(zone, w, r, m)

	default:
		m.SetRcode(r, dns.RcodeRefused)
	}

	// TSIG로 응답 서명
	if tsig != nil {
		m.SetTsig(tsig.Hdr.Name,
			tsig.Algorithm,
			300,
			time.Now().Unix())
	}

	err := w.WriteMsg(m)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (h *Handler) MsgAccept(dh dns.Header) dns.MsgAcceptAction {
	if isResponse := dh.Bits&_QR != 0; isResponse {
		return dns.MsgIgnore
	}

	// Don't allow dynamic updates, because then the sections can contain a whole bunch of RRs.
	opcode := int(dh.Bits>>11) & 0xF
	if opcode != dns.OpcodeQuery && opcode != dns.OpcodeUpdate {
		return dns.MsgRejectNotImplemented
	}

	if dh.Qdcount != 1 {
		return dns.MsgReject
	}
	// NOTIFY requests can have a SOA in the ANSWER section. See RFC 1996 Section 3.7 and 3.11.
	if dh.Ancount > 1 {
		return dns.MsgReject
	}
	// IXFR request could have one SOA RR in the NS section. See RFC 1995, section 3.
	if dh.Nscount > 1 {
		return dns.MsgReject
	}
	if dh.Arcount > 2 {
		return dns.MsgReject
	}
	return dns.MsgAccept
}

// SOA 쿼리를 위한 외부 DNS 조회 함수
func (h *Handler) queryUpstream(zone *domainZone, req *dns.Msg) (*dns.Msg, error) {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.Question = req.Question
	m.RecursionDesired = req.RecursionDesired
	m.AuthenticatedData = req.AuthenticatedData
	r, _, err := c.Exchange(m, zone.Upstream[0])
	return r, err

	//q := req.Question[0]
	//
	//current := q.Name
	//
	//var r *dns.Msg
	//for {
	//	var err error
	//
	//	nextDot := strings.Index(current, ".")
	//	if nextDot <= 0 {
	//		break
	//	}
	//
	//	m := new(dns.Msg)
	//	m.SetQuestion(current, q.Qtype)
	//	m.RecursionDesired = req.RecursionDesired
	//	m.AuthenticatedData = req.AuthenticatedData
	//
	//	r, _, err = c.Exchange(m, zone.Upstream[0])
	//	if err != nil {
	//		return nil, err
	//	}
	//	if len(r.Answer) > 0 {
	//		break
	//	}
	//
	//	current = current[nextDot+1:]
	//	if !req.RecursionDesired {
	//		break
	//	}
	//}
	//
	//return r, nil
}

func (h *Handler) writeTsigMsg(w dns.ResponseWriter, r *dns.Msg, m *dns.Msg) {
	tsig := r.Extra[len(r.Extra)-1].(*dns.TSIG)
	m.SetTsig(tsig.Hdr.Name,
		tsig.Algorithm,
		300,
		time.Now().Unix())

	err := w.WriteMsg(m)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (h *Handler) serveQuery(zone *domainZone, w dns.ResponseWriter, r *dns.Msg, m *dns.Msg) {
	log.Printf("ZONE[%s] Query to upstream: %+v", zone.Zone, r.Question)
	upstreamResp, err := h.queryUpstream(zone, r)
	if err != nil {
		log.Printf("\tFailed to query upstream: %v", err)
		m.SetRcode(r, dns.RcodeServerFailure)
	} else {
		log.Printf("\tRESP Rcode: %+v, Answers: %+v", upstreamResp.Rcode, upstreamResp.Answer)
		m.SetRcode(r, upstreamResp.Rcode)
		m.Answer = upstreamResp.Answer
		m.AuthenticatedData = upstreamResp.AuthenticatedData
		m.Authoritative = upstreamResp.Authoritative
	}
}

func (h *Handler) serveUpdate(zone *domainZone, w dns.ResponseWriter, r *dns.Msg, m *dns.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// Process the update message
	for _, update := range r.Ns {
		switch update.Header().Rrtype {
		case dns.TypeTXT:
			txt := update.(*dns.TXT)
			if len(txt.Txt) != 1 {
				m.SetRcode(r, dns.RcodeFormatError)
				log.Printf("invalid request: %+v", txt)
				return
			}

			log.Printf("Received TXT update for %s: %v", txt.Hdr.Name, txt.Txt)

			name := strings.TrimRight(txt.Hdr.Name, ".")

			records, err := zone.Client.GetRecord(ctx, strings.TrimRight(zone.Zone, "."), "TXT")
			if err != nil {
				m.SetRcode(r, dns.RcodeServerFailure)
				log.Printf("DNSEver GetRecord(%s, TXT) failed: %+v", name, err)
				return
			}

			var existingHost *dnsever.HostXml
			for _, host := range records.Result.Hosts {
				if host.Name == name {
					existingHost = &host
					break
				}
			}

			var result *dnsever.DNSEverXml
			var updateType string
			if existingHost != nil {
				updateType = "Update"
				result, err = zone.Client.UpdateRecord(ctx, existingHost.ID, "txt", txt.Txt[0], "", "")
			} else {
				updateType = "Add"
				result, err = zone.Client.AddRecord(ctx, name, "txt", txt.Txt[0], "", "")
			}
			if err != nil {
				m.SetRcode(r, dns.RcodeServerFailure)
				log.Printf("DNSEver %s Record(%s, TXT) failed: %+v", updateType, name, err)
				return
			} else {
				log.Printf("DNSEver %s Record(%s, TXT) result: %+v", updateType, name, result)
			}
			if result.Result.Code != 730 && result.Result.Code != 720 {
				// 730 : Add Success
				// 720 : Update Success
				m.SetRcode(r, dns.RcodeServerFailure)
				log.Printf("DNSEver %s Record(%s, TXT) failed: %d: %s", updateType, name, result.Result.Code, result.Result.Msg)
				return
			}
		default:
			log.Printf("Received unhandled update type: %d", update.Header().Rrtype)
		}
	}

	m.SetRcode(r, dns.RcodeSuccess)
}

func (h *Handler) validateTsig(w dns.ResponseWriter, tsig *dns.TSIG, allowEmpty bool) (zone *domainZone, ok bool) {
	// Verify TSIG
	if tsig != nil {
		if w.TsigStatus() != nil {
			log.Printf("authenticate failed: %+v", w.TsigStatus())
			return nil, false
		}
		auth := h.tsigAuth[tsig.Hdr.Name]
		if auth == nil {
			return nil, false
		}
		return auth.Zone, true
	} else {
		return nil, allowEmpty
	}
}

func (h *Handler) validateZone(found *domainZone, r *dns.Msg) *domainZone {
	for _, question := range r.Question {
		z := h.findZoneByName(question.Name)
		if z == nil {
			return nil
		} else if found == nil {
			found = z
		} else if found != z {
			return nil
		}
	}
	for _, ns := range r.Ns {
		z := h.findZoneByName(ns.Header().Name)
		if z == nil {
			return nil
		} else if found == nil {
			found = z
		} else if found != z {
			return nil
		}
	}
	return found
}

func (h *Handler) findZoneByName(name string) *domainZone {
	for _, domain := range h.domains {
		if name == domain.Zone || strings.HasSuffix(name, "."+domain.Zone) {
			return domain
		}
	}
	return nil
}
