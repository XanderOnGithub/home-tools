# Muscle Recovery Reference

Research-backed notes for estimating how long a muscle needs before it's ready to train again. Intended as a spec reference for the workout app.

> **In the app (2026-10-11, decision #49):** `tools/fitness/recovery.go` uses the per-muscle multipliers (§3) and a gentler novelty rule than §6 (sets of an exercise not done in 14 days count 1.5× instead of a 5-day floor). Effort (RIR), compound/eccentric flags, sleep and the readiness check (§7) need data the app doesn't log yet.

> **Bottom line:** Recovery time is driven mostly by **effort (proximity to failure)**, **volume**, and **novelty / eccentric emphasis**. Muscle *size* is a weak predictor. Treat every number here as a default range, and let logged performance override it.

---

## 1. Recovery tiers (default model)

| Tier | Exhaustion state | Default recovery window |
|---|---|---|
| **T1 — Light** | Sets stopped well short of failure (≈3+ RIR, or roughly half the possible reps) | **24–48 h** |
| **T2 — Hard** | Moderate volume, sets near failure (0–2 RIR) | **48–72 h** |
| **T3 — Wrecked** | High volume (≈6–8+ sets for the muscle) taken to failure | **72–96 h** |
| **T4 — Damage** | New exercise, eccentric-heavy work, or first session after a layoff | **5–14 days** (much faster on the 2nd exposure) |

### Evidence behind the tiers

- **Failure vs non-failure (Morán-Navarro 2017):** Going to failure every set (3×10 @ 10RM) caused a much bigger drop in performance than the same total volume without failure (6×5). The non-failure protocols recovered significantly faster at 24–48 h, and the authors concluded failure delays recovery by up to 24–48 h.
- **Velocity loss (Pareja-Blanco 2019):** Ending sets at 40% bar-speed loss was more fatiguing and slower to recover from than 20%. A lighter load (60% 1RM) was more fatiguing than a heavier one (80%).
- **Volume (Bartolomei 2017):** In trained men, 8×10 squats left isometric strength impaired at 72 h. With 8×3 heavy squats, strength was back to baseline.
- **Bench to failure (8 sets, trained men):** Peak strength returned at 96 h, but the ability to do repeated reps was *still* not recovered. A single max-effort test can say "recovered" too early.
- **Lower body to failure, trained men:** Performance needed at least 48 h to come back. Perceived recovery was better at 72 h.
- **Untrained, 4×10 @ 80% biceps:** Needed more than 72 h to fully recover.
- **Max eccentric work (Newham/Clarkson 1987):** Strength dropped about 50% and was only about 80% of baseline after 2 weeks. This is an extreme lab protocol, but it explains T4.

---

## 2. Modifiers

### Strong evidence: build these in
| Factor | Effect |
|---|---|
| **Proximity to failure** | Closer to failure means longer recovery. This is the biggest lever. |
| **Volume per muscle** | More hard sets means longer recovery. |
| **Novelty / layoff** | A new exercise, or one not done in more than 1–2 weeks, causes much more damage. |
| **Repeated bout effect** | The 2nd exposure causes far less damage. It can show up within about 3 days, and the protection fades between 1 and 2 weeks without the exercise. |
| **Eccentric / stretched emphasis** | Slow negatives, RDLs, Nordics, deep stretches: more damage, longer recovery. |
| **Compound vs isolation** | Multi-joint lifts were down about 1.7 reps at 24 h vs about 0.5 for single-joint lifts. Compounds also tax more than one muscle. |
| **Lighter load, high reps to failure** | More fatiguing than heavy low-rep work at the same proximity to failure. |

### Moderate evidence: optional tweaks
| Factor | Effect |
|---|---|
| **Age** | Older adults recover strength more slowly. |
| **Sleep** | One night without sleep cut muscle protein synthesis about 18%. Five nights of 4 h also lowered it. |
| **Sex** | Recovery was roughly similar between men and women. Men had a bigger immediate drop. Probably not worth modeling. |

### Weak or misleading: don't rely on these
| Factor | Why |
|---|---|
| **Soreness (DOMS)** | Correlates poorly with strength recovery. Reduced pain doesn't mean strength is back. |
| **Muscle size** | See section 3. |

---

## 3. Per-muscle differences

**The intuition that bigger muscles are slower isn't supported. The data leans the opposite way.**

- **Arms vs legs:** With the same relative eccentric load, arm muscles take more damage and recover more slowly than the quads. A 9-muscle study also found more damage in arm muscles than in leg muscles.
- **Why:** Leg muscles are used all day and have a different architecture, which makes them more fatigue-resistant. The biceps also does more eccentric work per unit of muscle.
- **Hamstrings vs quads:** Hamstrings are more susceptible to damage than quads.
- **Under normal training, differences shrink:** In well-trained lifters doing 4 sets to failure at 80%, squat, bench and deadlift recovered on the **same** timeline.
- **Mixed findings:** In some groups, bench recovered slower than squat. In recreational lifters, bench and deadlift had the most people still not recovered at 48 h.

### Suggested per-muscle multipliers (keep them small, let users edit)

| Muscle group | Multiplier (applied mainly after eccentric-heavy work) |
|---|---|
| Quads, calves, glutes | 0.9× |
| Chest, back, shoulders | 1.0× |
| Hamstrings | 1.1× |
| Biceps, triceps | 1.1× |

> These are design defaults, not study values. Effort, volume and novelty matter far more than which muscle it is.

---

## 4. The different "recovery clocks"

These run on separate timelines. Don't conflate them in the UI.

| Clock | Typical timeline | Notes |
|---|---|---|
| **Glycogen** (fuel) | Mostly restored in ~6 h | Fastest; rarely the limiter for lifting |
| **Muscle protein synthesis** (growth signal) | Trained: elevated ~24 h, near baseline by ~36 h. Untrained: still up at 48 h | Supports training a muscle about 2×/week |
| **Neuromuscular performance** (can you lift well) | 24–96 h depending on tier | **The one the app should track** |
| **Muscle damage** | Days to 2 weeks for T4 | Mostly matters for novel or eccentric work |
| **Soreness** | Peaks ~24–72 h | Unreliable signal |

**Frequency note:** With weekly volume held equal, training frequency doesn't meaningfully change muscle growth (Schoenfeld 2019). Higher frequency helps mostly because it lets you fit in more good-quality volume.

---

## 5. Data to log (minimum useful set)

Per set:
- `exercise_id`
- `weight`
- `reps`
- `rir` (reps in reserve, 0–5), or a `to_failure` boolean as a simpler alternative

Per exercise (static metadata):
- `primary_muscles[]`, `secondary_muscles[]`
- `is_compound` (bool)
- `eccentric_emphasis` (bool), e.g. RDL, Nordic, deficit, slow-negative variants

Derived:
- `last_performed_at` per exercise (to detect novelty or layoff)
- hard sets per muscle per session (sets with RIR ≤ 3)

Optional:
- `sleep_quality` (1–5) per session
- user birthday (age modifier)

---

## 6. Suggested estimation logic

```text
for each muscle trained in the session:
    hard_sets  = count(sets where muscle is primary AND rir <= 3)
               + 0.5 * count(sets where muscle is secondary AND rir <= 3)
    avg_rir    = average RIR of those sets

    # base tier
    if avg_rir >= 3:                         base = 36h   # T1
    elif hard_sets >= 6 and avg_rir <= 1:    base = 84h   # T3
    else:                                    base = 60h   # T2

    # modifiers
    if any exercise for this muscle is new or not done in 14+ days:
        base = max(base, 120h)               # T4 floor (first exposure)
    if any exercise is eccentric_emphasis:   base *= 1.2
    if any exercise is_compound:             base *= 1.1
    base *= muscle_multiplier[muscle]        # small, editable
    if sleep_quality <= 2:                   base *= 1.1

    ready_at = session_end + base
```

All constants are tunable defaults derived from the ranges above. Show the user a **range** ("likely ready in 2–3 days"), not a precise countdown.

---

## 7. Readiness check (overrides the estimate)

The best recovery signal is performance:

1. When the user does the first working set of an exercise, compare it with the last session at the same weight.
2. **Reps ≥ last time:** recovered. Optionally shorten that muscle's future estimates slightly.
3. **Reps down by 1:** borderline. Suggest keeping volume or easing off on proximity to failure.
4. **Reps down by 2+:** not recovered. Suggest reducing sets or switching to a different muscle.
5. Over time, learn a personal adjustment factor per muscle from these comparisons. Differences between people are larger than differences between muscles.

---

## 8. Caveats

- Most studies have 10–20 participants and are mostly young men.
- They use lab measures (isometric strength, jump height, bar velocity), not normal gym sets.
- The direction of the findings is solid; the exact hours are approximate.
- Eccentric damage studies use extreme protocols on untrained people. Normal training rarely hits T4 except with new exercises or after a layoff.

---

## References

- Morán-Navarro et al. (2017). Time course of recovery following resistance training leading or not to failure. *Eur J Appl Physiol* 117:2387–2399. https://pubmed.ncbi.nlm.nih.gov/28965198/
- Pareja-Blanco et al. (2019). Time course of recovery following resistance exercise with different loading magnitudes and velocity loss in the set. *Sports* 7(3):59. https://doaj.org/article/7e02ce43752c47b194ac37dd4fed5c7b
- Bartolomei et al. (2017). Comparison of the recovery response from high-intensity and high-volume resistance exercise in trained men. *Eur J Appl Physiol* 117:1287–1298. https://stars.library.ucf.edu/scopus2015/5190
- Bench press to failure, peak torque vs total work recovery. https://paulogentil.com/?p=339
- Recovery interval after lower-limb training to failure. https://lida.sport-iat.de/ta/Record/4062986
- Radaelli et al. Time course of strength and echo intensity recovery after resistance exercise in women. *J Strength Cond Res* 26(9). https://www.bisp-surf.de/Record/PU201301000216
- Newham, Jones & Clarkson (1987). Repeated high-force eccentric exercise: effects on muscle pain and damage. https://paulogentil.com/pdf/Repeated%20high-force%20eccentric%20exercise%20-%20effects%20on%20muscle%20pain%20and%20damage.pdf
- Repeated bout effect in resistance-trained men (eccentric bench). https://independent.academia.edu/CharlesLopes1
- Compression garments and the repeated bout effect (RBE onset). https://pmc.ncbi.nlm.nih.gov/articles/PMC10988506
- Chen, Nosaka & Chen (2012). Muscle damage protection remains for 2 weeks but not 3. https://pubmed.ncbi.nlm.nih.gov/21611825/
- Comparison between leg and arm eccentric exercises of the same relative intensity. https://ir.lib.uth.gr/xmlui/handle/11615/28650?show=full
- Chen et al. Damage and the repeated bout effect of arm, leg, and trunk muscles. https://ro.ecu.edu.au/ecuworkspost2013/6052
- Chen et al. (2011). Comparison in eccentric exercise-induced muscle damage among four limb muscles. https://paulogentil.com/pdf/TREINO%20DE%20FORC%CC%A7A/Outras%20adaptac%CC%A7o%CC%83es/Comparison%20in%20eccentric%20exercise-induced%20muscle%20damage%20among%20four%20limb%20muscles.pdf
- Overview of muscle fatigue differences between maximal eccentric and concentric exercise. https://ro.ecu.edu.au/ecuworks2022-2026/2796/
- Time course of recovery is similar for the back squat, bench press, and deadlift in well-trained males. https://lida.sport-iat.de/ta/Record/4056507?lng=en
- Korak, Green & O'Neal (2015). Resistance training recovery: single vs multi-joint, upper vs lower body. https://digitalcommons.wku.edu/ijes/vol8/iss1/10
- Mitchell et al. (2014). Upper- and lower-body strength training recovery in male adolescent athletes. https://coachsci.sdsu.edu/csa/vol216/mitchell.htm
- Men and women experience similar muscle damage after traditional resistance training. https://digitalcommons.usu.edu/hper_facpub/153
- Dedrick & Clarkson (1990). Eccentric exercise in young and older women. https://www.bisp-surf.de/Record/PU199005043373
- Acute sleep deprivation and muscle protein synthesis (2021). https://vuir.vu.edu.au/43098/
- Saner et al. (2020). Sleep restriction and myofibrillar protein synthesis. https://experts.mcmaster.ca/scholarly-works/1695629
- Denegar & Perrin. DOMS treatment and strength recovery. https://www.bisp-surf.de/Record/ftpubmed:oai:pubmedcentral.nih.gov:1317247
- MacDougall et al. (1995). Time course for elevated muscle protein synthesis. https://pubmed.ncbi.nlm.nih.gov/8563679/
- Phillips et al. (1997). Mixed muscle protein synthesis and breakdown after resistance exercise. https://experts.mcmaster.ca/scholarly-works/184576
- Pascoe et al. (1993). Glycogen resynthesis after resistive exercise. https://pubmed.ncbi.nlm.nih.gov/8455450/
- Schoenfeld, Grgic & Krieger (2019). Training frequency meta-analysis (summary table). https://pmc.ncbi.nlm.nih.gov/articles/PMC9302196/table/T1
